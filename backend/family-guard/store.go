package main

import (
    "context"
    "database/sql"
    "embed"
    "encoding/json"
    "fmt"
    "os"
    "sort"
    "strconv"
    "strings"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStore struct { db *sql.DB }

//go:embed migrations/*.sql
var migrationFiles embed.FS

func OpenSQLStore(ctx context.Context) (*SQLStore, error) {
    dsn := os.Getenv("FTN_FAMILY_GUARD_DATABASE_URL")
    if dsn == "" { return nil, nil }
    db, err := sql.Open("pgx", dsn)
    if err != nil { return nil, fmt.Errorf("open postgres: %w", err) }
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(30 * time.Minute)
    s := &SQLStore{db: db}
    if err := s.Ping(ctx); err != nil { _ = db.Close(); return nil, fmt.Errorf("postgres ping: %w", err) }
    if err := s.RunMigrations(ctx); err != nil { _ = db.Close(); return nil, err }
    return s, nil
}

func (s *SQLStore) RunMigrations(ctx context.Context) error {
    if s == nil || s.db == nil { return fmt.Errorf("database is not configured") }
    ctx, cancel := context.WithTimeout(ctx, 15*time.Second); defer cancel()
    if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS fg_schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil { return fmt.Errorf("migration table: %w", err) }
    rows, err := s.db.QueryContext(ctx, `SELECT version FROM fg_schema_migrations`)
    if err != nil { return fmt.Errorf("read migrations: %w", err) }
    applied := map[string]bool{}
    for rows.Next() { var v string; if err := rows.Scan(&v); err != nil { rows.Close(); return err }; applied[v] = true }
    if err := rows.Err(); err != nil { rows.Close(); return err }; rows.Close()
    entries, err := migrationFiles.ReadDir("migrations")
    if err != nil { return fmt.Errorf("read migrations: %w", err) }
    sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
    for _, e := range entries {
        if e.IsDir() || applied[e.Name()] { continue }
        sqlBytes, err := migrationFiles.ReadFile("migrations/" + e.Name()); if err != nil { return err }
        if _, err := s.db.ExecContext(ctx, string(sqlBytes)); err != nil { return fmt.Errorf("apply %s: %w", e.Name(), err) }
        if _, err := s.db.ExecContext(ctx, `INSERT INTO fg_schema_migrations(version) VALUES($1) ON CONFLICT DO NOTHING`, e.Name()); err != nil { return fmt.Errorf("record %s: %w", e.Name(), err) }
    }
    return nil
}

func (s *SQLStore) Ping(ctx context.Context) error {
    if s == nil || s.db == nil { return fmt.Errorf("database is not configured") }
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second); defer cancel()
    return s.db.PingContext(ctx)
}

func (s *SQLStore) Close() error { if s == nil || s.db == nil { return nil }; return s.db.Close() }

func (s *SQLStore) ActivePolicy(ctx context.Context) (map[string]any, error) {
    if s == nil || s.db == nil { return nil, fmt.Errorf("database is not configured") }
    var raw []byte
    err := s.db.QueryRowContext(ctx, `SELECT policy FROM fg_policies WHERE active=true ORDER BY version DESC LIMIT 1`).Scan(&raw)
    if err == sql.ErrNoRows { return nil, nil }
    if err != nil { return nil, err }
    var p map[string]any
    if err := json.Unmarshal(raw, &p); err != nil { return nil, err }
    return p, nil
}

func (s *SQLStore) SavePolicy(ctx context.Context, policy map[string]any) error {
    if s == nil || s.db == nil { return fmt.Errorf("database is not configured") }
    version, ok := numericVersion(policy["version"])
    if !ok || version < 1 { return fmt.Errorf("invalid policy version") }
    profile, _ := policy["profile"].(string)
    if strings.TrimSpace(profile) == "" { profile = "FAMILY" }
    raw, err := json.Marshal(policy); if err != nil { return err }
    tx, err := s.db.BeginTx(ctx, nil); if err != nil { return err }
    defer tx.Rollback()
    if _, err = tx.ExecContext(ctx, `UPDATE fg_policies SET active=false WHERE active=true`); err != nil { return err }
    _, err = tx.ExecContext(ctx, `INSERT INTO fg_policies(version,profile,policy,active) VALUES($1,$2,$3::jsonb,true) ON CONFLICT(version) DO UPDATE SET profile=EXCLUDED.profile, policy=EXCLUDED.policy, active=true, created_at=now()`, version, profile, raw)
    if err != nil { return err }
    return tx.Commit()
}

func (s *SQLStore) ListDevices(ctx context.Context) ([]map[string]any, error) {
    if s == nil || s.db == nil { return nil, fmt.Errorf("database is not configured") }
    rows, err := s.db.QueryContext(ctx, `SELECT id,name,profile,status,policy_version,enrolled_at,last_seen_at FROM fg_devices ORDER BY id`)
    if err != nil { return nil, err }
    defer rows.Close()
    out := []map[string]any{}
    for rows.Next() {
        var id,name,profile,status string; var version int64; var enrolled time.Time; var last sql.NullTime
        if err := rows.Scan(&id,&name,&profile,&status,&version,&enrolled,&last); err != nil { return nil, err }
        d := map[string]any{"id":id,"name":name,"profile":profile,"status":status,"policy_version":version,"enrolled_at":enrolled}
        if last.Valid { d["last_seen_at"] = last.Time }
        out = append(out,d)
    }
    return out, rows.Err()
}

func (s *SQLStore) UpsertDevice(ctx context.Context, device map[string]any) error {
    if s == nil || s.db == nil { return fmt.Errorf("database is not configured") }
    id, _ := device["id"].(string); id = strings.TrimSpace(id); if id == "" { return fmt.Errorf("device id is required") }
    name, _ := device["name"].(string); if name == "" { name = "Android Device" }
    profile, _ := device["profile"].(string); if profile == "" { profile = "FAMILY" }
    status, _ := device["status"].(string); if status == "" { status = "online" }
    version, _ := numericVersion(device["policy_version"]); if version < 1 { version = 1 }
    _, err := s.db.ExecContext(ctx, `INSERT INTO fg_devices(id,name,profile,status,policy_version,enrolled_at,last_seen_at) VALUES($1,$2,$3,$4,$5,now(),now()) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,profile=EXCLUDED.profile,status=EXCLUDED.status,policy_version=EXCLUDED.policy_version,last_seen_at=now()`, id,name,profile,status,version)
    return err
}

func (s *SQLStore) ListRules(ctx context.Context, kind string) ([]map[string]any, error) {
    if s == nil || s.db == nil { return nil, fmt.Errorf("database is not configured") }
    table := ruleTable(kind); if table == "" { return nil, fmt.Errorf("invalid rule kind") }
    rows, err := s.db.QueryContext(ctx, `SELECT id,value,action,profile,created_at FROM `+table+` ORDER BY id`)
    if err != nil { return nil, err }
    defer rows.Close()
    out := []map[string]any{}
    for rows.Next() {
        var id int64; var value,action string; var profile sql.NullString; var created time.Time
        if err := rows.Scan(&id,&value,&action,&profile,&created); err != nil { return nil, err }
        v := map[string]any{"id":id,"value":value,"action":action,"created_at":created}; if profile.Valid { v["profile"] = profile.String }
        out = append(out,v)
    }
    return out, rows.Err()
}

func (s *SQLStore) AddRule(ctx context.Context, kind string, rule map[string]any) error {
    if s == nil || s.db == nil { return fmt.Errorf("database is not configured") }
    table := ruleTable(kind); if table == "" { return fmt.Errorf("invalid rule kind") }
    value, _ := rule["value"].(string); action, _ := rule["action"].(string); profile, _ := rule["profile"].(string)
    value = strings.TrimSpace(value); action = strings.TrimSpace(action)
    if value == "" || (action != "allow" && action != "block") { return fmt.Errorf("invalid rule") }
    _, err := s.db.ExecContext(ctx, `INSERT INTO `+table+`(value,action,profile) VALUES($1,$2,NULLIF($3,''))`, value,action,profile)
    return err
}

func ruleTable(kind string) string { switch kind { case "domain": return "fg_domain_rules"; case "ip": return "fg_ip_rules"; default: return "" } }

func numericVersion(v any) (int64, bool) {
    switch x := v.(type) {
    case int: return int64(x), true
    case int64: return x, true
    case float64: return int64(x), x == float64(int64(x))
    case json.Number: n,e:=x.Int64(); return n,e==nil
    case string: n,e:=strconv.ParseInt(strings.TrimSpace(x),10,64); return n,e==nil
    default: return 0,false
    }
}

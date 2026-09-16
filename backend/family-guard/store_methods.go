package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
)

func (s *SQLStore) UpsertDevice(ctx context.Context, d map[string]any) error {
    id, _ := d["id"].(string); name, _ := d["name"].(string); profile, _ := d["profile"].(string); status, _ := d["status"].(string)
    if id == "" { return fmt.Errorf("device id is required") }; if name == "" { name = "Android Device" }; if profile == "" { profile = "FAMILY" }; if status == "" { status = "offline" }
    version := int64(1); if v, ok := d["policy_version"].(float64); ok && v >= 1 { version = int64(v) }
    _, err := s.db.ExecContext(ctx, `INSERT INTO fg_devices(id,name,profile,status,policy_version,last_seen_at) VALUES($1,$2,$3,$4,$5,now()) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,profile=EXCLUDED.profile,status=EXCLUDED.status,policy_version=EXCLUDED.policy_version,last_seen_at=now()`, id,name,profile,status,version)
    return err
}

func (s *SQLStore) ListDevices(ctx context.Context) ([]map[string]any, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT id,name,profile,status,policy_version,enrolled_at,last_seen_at FROM fg_devices ORDER BY id`); if err != nil { return nil, err }; defer rows.Close()
    out := []map[string]any{}
    for rows.Next() { var id,name,profile,status string; var version int64; var enrolled time.Time; var last sql.NullTime; if err:=rows.Scan(&id,&name,&profile,&status,&version,&enrolled,&last);err!=nil{return nil,err}; d:=map[string]any{"id":id,"name":name,"profile":profile,"status":status,"policy_version":version,"enrolled_at":enrolled}; if last.Valid { d["last_seen_at"]=last.Time }; out=append(out,d) }
    return out, rows.Err()
}

func (s *SQLStore) SavePolicy(ctx context.Context, policy map[string]any) error {
    version, ok := policy["version"].(float64); if !ok || version < 1 { return fmt.Errorf("policy version is required") }; profile,_:=policy["profile"].(string); if profile=="" { profile="FAMILY" }; b,err:=json.Marshal(policy); if err!=nil{return err}
    tx,err:=s.db.BeginTx(ctx,nil);if err!=nil{return err};defer tx.Rollback()
    if _,err=tx.ExecContext(ctx,`UPDATE fg_policies SET active=FALSE WHERE active`);err!=nil{return err}
    if _,err=tx.ExecContext(ctx,`INSERT INTO fg_policies(version,profile,policy,active) VALUES($1,$2,$3,TRUE) ON CONFLICT(version) DO UPDATE SET profile=EXCLUDED.profile,policy=EXCLUDED.policy,active=TRUE`,int64(version),profile,b);err!=nil{return err}
    return tx.Commit()
}

func (s *SQLStore) ActivePolicy(ctx context.Context) (map[string]any, error) {
    var raw []byte; err:=s.db.QueryRowContext(ctx,`SELECT policy FROM fg_policies WHERE active LIMIT 1`).Scan(&raw); if err==sql.ErrNoRows{return nil,nil}; if err!=nil{return nil,err}; var p map[string]any; if err:=json.Unmarshal(raw,&p);err!=nil{return nil,err}; return p,nil
}

func (s *SQLStore) AddRule(ctx context.Context, kind string, rule map[string]any) error {
    value,_:=rule["value"].(string); action,_:=rule["action"].(string); profile,_:=rule["profile"].(string); if kind=="domain" {_,err:=s.db.ExecContext(ctx,`INSERT INTO fg_domain_rules(value,action,profile) VALUES($1,$2,NULLIF($3,''))`,value,action,profile);return err}; if kind=="ip" {_,err:=s.db.ExecContext(ctx,`INSERT INTO fg_ip_rules(value,action,profile) VALUES($1,$2,NULLIF($3,''))`,value,action,profile);return err};return fmt.Errorf("unknown rule kind %q",kind)
}

func (s *SQLStore) ListRules(ctx context.Context, kind string) ([]map[string]any,error) {
    var rows *sql.Rows; var err error; if kind=="domain" {rows,err=s.db.QueryContext(ctx,`SELECT id,value,action,COALESCE(profile,''),created_at FROM fg_domain_rules ORDER BY id`)} else if kind=="ip" {rows,err=s.db.QueryContext(ctx,`SELECT id,value,action,COALESCE(profile,''),created_at FROM fg_ip_rules ORDER BY id`)} else{return nil,fmt.Errorf("unknown rule kind %q",kind)}; if err!=nil{return nil,err};defer rows.Close();out:=[]map[string]any{};for rows.Next(){var id int64;var value,action,profile string;var created time.Time;if err:=rows.Scan(&id,&value,&action,&profile,&created);err!=nil{return nil,err};out=append(out,map[string]any{"id":id,"value":value,"action":action,"profile":profile,"created_at":created})};return out,rows.Err()
}

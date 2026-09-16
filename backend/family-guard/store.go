package main

import (
    "context"
    "database/sql"
    "embed"
    "fmt"
    "os"
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

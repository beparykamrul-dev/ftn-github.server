package main

import (
    "context"
    "database/sql"
    "embed"
    "fmt"
    "strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func runMigrations(ctx context.Context, db *sql.DB) error {
    if db == nil { return fmt.Errorf("database is not configured") }
    if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS fg_schema_migrations (version BIGINT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
        return err
    }

    const version int64 = 1
    var applied bool
    if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM fg_schema_migrations WHERE version=$1)`, version).Scan(&applied); err != nil {
        return err
    }
    if applied { return nil }

    b, err := migrationFS.ReadFile("migrations/001_family_guard.sql")
    if err != nil { return err }
    statements := strings.Split(string(b), ";")

    tx, err := db.BeginTx(ctx, nil)
    if err != nil { return err }
    defer tx.Rollback()
    for _, statement := range statements {
        statement = strings.TrimSpace(statement)
        if statement == "" || statement == "BEGIN" || statement == "COMMIT" { continue }
        if _, err := tx.ExecContext(ctx, statement); err != nil { return err }
    }
    if _, err := tx.ExecContext(ctx, `INSERT INTO fg_schema_migrations(version) VALUES ($1)`, version); err != nil { return err }
    return tx.Commit()
}

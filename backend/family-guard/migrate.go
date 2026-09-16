package main

import (
    "context"
    "embed"
    "fmt"
    "strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func runMigrations(ctx context.Context, db interface {
    ExecContext(context.Context, string, ...any) (any, error)
}) error {
    // Kept as a small contract helper; the concrete implementation is in runSQLMigrations.
    return fmt.Errorf("unsupported migration database")
}

// migrationDB is the minimal database/sql contract used by the runner.
type migrationDB interface {
    ExecContext(context.Context, string, ...any) (sqlResult, error)
}

type sqlResult interface{}

func migrationStatements() ([]string, error) {
    b, err := migrationFS.ReadFile("migrations/001_family_guard.sql")
    if err != nil { return nil, err }
    parts := strings.Split(string(b), ";")
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" && p != "BEGIN" && p != "COMMIT" { out = append(out, p) }
    }
    return out, nil
}

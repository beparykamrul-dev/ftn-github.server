package main

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "time"
)

// SQLStore is the persistence boundary. The HTTP state layer can use it when a
// PostgreSQL DSN is provisioned; credentials are never stored in Git.
type SQLStore struct { db *sql.DB }

func OpenSQLStore(ctx context.Context) (*SQLStore, error) {
    dsn := os.Getenv("FTN_FAMILY_GUARD_DATABASE_URL")
    if dsn == "" { return nil, nil }
    // Driver registration is intentionally owned by the deployment build.
    // This package only defines the storage contract and lifecycle.
    return nil, fmt.Errorf("PostgreSQL DSN configured but no SQL driver is linked")
}

func (s *SQLStore) Ping(ctx context.Context) error {
    if s == nil || s.db == nil { return fmt.Errorf("database is not configured") }
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second); defer cancel()
    return s.db.PingContext(ctx)
}

func (s *SQLStore) Close() error {
    if s == nil || s.db == nil { return nil }
    return s.db.Close()
}

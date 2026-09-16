# PostgreSQL activation

The Family Guard service uses `FTN_FAMILY_GUARD_DATABASE_URL` for PostgreSQL. Runtime credentials stay outside Git.

The repository migration is `migrations/001_family_guard.sql` and is additive/non-destructive.

Before production start, the deployment image must include the `github.com/jackc/pgx/v5` module and the SQL driver registration used by `store.go`.

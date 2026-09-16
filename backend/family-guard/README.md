# FTN Family Guard backend

## Runtime

- HTTP: `FTN_FAMILY_GUARD_ADDR` (default `:8095`)
- PostgreSQL: `FTN_FAMILY_GUARD_DATABASE_URL`
- API authentication: `FTN_FAMILY_GUARD_API_TOKEN`
- CORS: `FTN_FAMILY_GUARD_ALLOWED_ORIGIN`

## PostgreSQL activation

1. Provision a dedicated PostgreSQL database outside Git.
2. Set `FTN_FAMILY_GUARD_DATABASE_URL` in the service environment.
3. Apply `migrations/001_family_guard.sql` with a migration runner or `psql`.
4. Start the backend only after the schema is available.
5. Verify `/healthz` and `/api/v1/health/summary`.

The migration is additive and does not drop existing tables/data.

## Privacy boundary

Persist only device/policy/rule/aggregate-health metadata. Do not persist raw DNS questions, packet payloads, credentials, or private resolver addresses.

## Resolver policy

FTN remains preferred. Cloudflare Family (`1.1.1.1` / `1.0.0.1`) is not hard-blocked; it is selectable only when the active resolver policy permits it. DNSSEC remains required.

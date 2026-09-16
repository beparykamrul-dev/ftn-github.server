BEGIN;

CREATE TABLE IF NOT EXISTS fg_devices (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  profile TEXT NOT NULL DEFAULT 'FAMILY',
  status TEXT NOT NULL DEFAULT 'offline',
  policy_version BIGINT NOT NULL DEFAULT 1,
  enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS fg_policies (
  version BIGINT PRIMARY KEY,
  profile TEXT NOT NULL,
  policy JSONB NOT NULL,
  active BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS fg_one_active_policy ON fg_policies(active) WHERE active;

CREATE TABLE IF NOT EXISTS fg_domain_rules (
  id BIGSERIAL PRIMARY KEY,
  value TEXT NOT NULL,
  action TEXT NOT NULL CHECK (action IN ('allow','block')),
  profile TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fg_ip_rules (
  id BIGSERIAL PRIMARY KEY,
  value TEXT NOT NULL,
  action TEXT NOT NULL CHECK (action IN ('allow','block')),
  profile TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fg_usage_daily (
  day DATE NOT NULL,
  device_id TEXT REFERENCES fg_devices(id) ON DELETE CASCADE,
  allowed BIGINT NOT NULL DEFAULT 0,
  blocked BIGINT NOT NULL DEFAULT 0,
  failed BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (day, device_id)
);

CREATE TABLE IF NOT EXISTS fg_resolver_health (
  resolver_id TEXT PRIMARY KEY,
  status TEXT NOT NULL,
  dnssec_required BOOLEAN NOT NULL DEFAULT TRUE,
  latency_ms INTEGER,
  checked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fg_audit_events (
  id BIGSERIAL PRIMARY KEY,
  event_type TEXT NOT NULL,
  actor_id TEXT,
  device_id TEXT,
  request_id TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- No raw DNS questions, DNS payloads, credentials, or private resolver addresses are persisted.

COMMIT;

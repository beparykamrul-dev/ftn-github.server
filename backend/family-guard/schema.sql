-- FTN Family Guard persistent model.
-- Raw DNS queries, payloads, credentials and private resolver addresses are intentionally not stored.
CREATE TABLE IF NOT EXISTS family_devices (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  profile TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'offline',
  policy_version BIGINT NOT NULL DEFAULT 1,
  enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS family_policies (
  id BIGSERIAL PRIMARY KEY,
  version BIGINT NOT NULL UNIQUE,
  profile TEXT NOT NULL,
  dns_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  dnssec_required BOOLEAN NOT NULL DEFAULT TRUE,
  encrypted_dns_required BOOLEAN NOT NULL DEFAULT TRUE,
  ftn_resolver_allowed BOOLEAN NOT NULL DEFAULT TRUE,
  cloudflare_family_allowed BOOLEAN NOT NULL DEFAULT FALSE,
  public_fallback_allowed BOOLEAN NOT NULL DEFAULT FALSE,
  policy_json JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dns_domain_rules (
  id BIGSERIAL PRIMARY KEY,
  domain TEXT NOT NULL,
  action TEXT NOT NULL CHECK (action IN ('allow','block')),
  profile TEXT,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dns_domain_rules_domain ON dns_domain_rules(domain);

CREATE TABLE IF NOT EXISTS dns_ip_rules (
  id BIGSERIAL PRIMARY KEY,
  cidr TEXT NOT NULL,
  action TEXT NOT NULL CHECK (action IN ('allow','block')),
  profile TEXT,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dns_usage_daily (
  day DATE NOT NULL,
  device_id TEXT NOT NULL REFERENCES family_devices(id) ON DELETE CASCADE,
  allowed BIGINT NOT NULL DEFAULT 0,
  blocked BIGINT NOT NULL DEFAULT 0,
  failed BIGINT NOT NULL DEFAULT 0,
  top_categories JSONB NOT NULL DEFAULT '{}'::jsonb,
  PRIMARY KEY(day, device_id)
);

CREATE TABLE IF NOT EXISTS resolver_health (
  resolver_id TEXT PRIMARY KEY,
  status TEXT NOT NULL,
  latency_ms INTEGER,
  dnssec_status TEXT NOT NULL,
  checked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS family_audit_events (
  id BIGSERIAL PRIMARY KEY,
  event_type TEXT NOT NULL,
  actor_type TEXT NOT NULL,
  actor_id TEXT,
  device_id TEXT,
  policy_version BIGINT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

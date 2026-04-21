-- ============================================================================
-- DNSC Microservice – Database initialization
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS core;

CREATE TABLE IF NOT EXISTS core.domains (
    id UUID PRIMARY KEY,
    value TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('whitelist', 'blacklist', 'pending', 'rejected')),
    description TEXT,
    pnrisc_last_synced_at TIMESTAMPTZ,
    pnrisc_sync_status TEXT,
    pnrisc_remote_id TEXT,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Baze create înainte de coloana description: adaugă fără a recrea tabela
ALTER TABLE core.domains ADD COLUMN IF NOT EXISTS description TEXT;

CREATE OR REPLACE FUNCTION core.domains_touch_last_updated()
RETURNS TRIGGER AS $$
BEGIN
  IF (OLD.value IS DISTINCT FROM NEW.value
      OR OLD.type IS DISTINCT FROM NEW.type
      OR OLD.status IS DISTINCT FROM NEW.status
      OR OLD.description IS DISTINCT FROM NEW.description) THEN
    NEW.last_updated := now();
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_domains_last_updated ON core.domains;
CREATE TRIGGER trg_domains_last_updated
BEFORE UPDATE ON core.domains
FOR EACH ROW
EXECUTE PROCEDURE core.domains_touch_last_updated();

CREATE TABLE IF NOT EXISTS core.domain_records (
    id UUID PRIMARY KEY,
    domain_id UUID NOT NULL REFERENCES core.domains(id) ON DELETE CASCADE,
    ticket_id TEXT NOT NULL,
    description TEXT,
    tags TEXT[],
    date TIMESTAMPTZ NOT NULL,
    source TEXT,
    last_successful_sync_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_domains_value ON core.domains(value);
CREATE INDEX IF NOT EXISTS idx_domains_type ON core.domains(type);
CREATE INDEX IF NOT EXISTS idx_domains_status ON core.domains(status);
CREATE INDEX IF NOT EXISTS idx_domain_records_domain_id ON core.domain_records(domain_id);
CREATE INDEX IF NOT EXISTS idx_domain_records_date ON core.domain_records(date);


CREATE TABLE IF NOT EXISTS core.domain_status (
    id UUID PRIMARY KEY,
    domain_id UUID NOT NULL REFERENCES core.domains(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('whitelist', 'blacklist', 'pending', 'rejected')),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by TEXT NOT NULL,
    notes TEXT
);

CREATE INDEX IF NOT EXISTS idx_domain_status_domain_id ON core.domain_status(domain_id);
CREATE INDEX IF NOT EXISTS idx_domain_status_changed_at ON core.domain_status(changed_at);

CREATE TABLE IF NOT EXISTS core.whitelist_requests (
    id UUID PRIMARY KEY,
    domain_id UUID NOT NULL REFERENCES core.domains(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    address TEXT NOT NULL,
    phone TEXT NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_whitelist_requests_domain_id ON core.whitelist_requests(domain_id);
CREATE INDEX IF NOT EXISTS idx_whitelist_requests_created_at ON core.whitelist_requests(created_at);

-- Failed RTIR ticket imports (empty IOC, fetch/extract/upsert errors); cleared on successful sync.
CREATE TABLE IF NOT EXISTS core.rtir_import_errors (
    id UUID PRIMARY KEY,
    ticket_id TEXT NOT NULL UNIQUE,
    source TEXT NOT NULL DEFAULT 'rtir-sync',
    error_message TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    last_sync_try_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rtir_import_errors_last_try ON core.rtir_import_errors(last_sync_try_at);

-- ---------------------------------------------------------------------------
-- Etichete (UI + domenii); listă administrată în DB
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS core.tag (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    value TEXT NOT NULL UNIQUE,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_tag_sort_order ON core.tag(sort_order);

INSERT INTO core.tag (value, sort_order) VALUES
    ('malware', 10),
    ('phishing', 20),
    ('brute-force', 30),
    ('ransomware', 40),
    ('c2', 50),
    ('apt', 60),
    ('ddos', 70),
    ('dns-tunnel', 80),
    ('exfiltration', 90),
    ('port-scan', 100),
    ('social-engineering', 110),
    ('ssh', 120),
    ('critical', 130),
    ('false-positive', 140),
    ('verified', 150),
    ('internal', 160)
ON CONFLICT (value) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Auth: utilizatori și sesiuni (cookie session_id pe server)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS core.app_user (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION core.app_user_touch_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_app_user_updated_at ON core.app_user;
CREATE TRIGGER trg_app_user_updated_at
BEFORE UPDATE ON core.app_user
FOR EACH ROW
EXECUTE PROCEDURE core.app_user_touch_updated_at();

CREATE TABLE IF NOT EXISTS core.user_session (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES core.app_user(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_session_user_id ON core.user_session(user_id);
CREATE INDEX IF NOT EXISTS idx_user_session_expires_at ON core.user_session(expires_at);

-- Utilizator inițial (schimbă parola în producție)
INSERT INTO core.app_user (id, username, password, is_active)
VALUES ('00000000-0000-0000-0000-000000000001', 'admin', 'admin', true)
ON CONFLICT (username) DO NOTHING;

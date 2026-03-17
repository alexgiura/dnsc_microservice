-- ============================================================================
-- DNSC Microservice – Database initialization
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS core;

CREATE TABLE IF NOT EXISTS core.domains (
    id UUID PRIMARY KEY,
    value TEXT NOT NULL,
    type TEXT NOT NULL,
    whitelist BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS core.domain_records (
    id UUID PRIMARY KEY,
    domain_id UUID NOT NULL REFERENCES core.domains(id) ON DELETE CASCADE,
    ticket_id TEXT NOT NULL,
    description TEXT,
    tags TEXT[],
    date TIMESTAMPTZ NOT NULL,
    source TEXT
);

CREATE INDEX IF NOT EXISTS idx_domains_value ON core.domains(value);
CREATE INDEX IF NOT EXISTS idx_domains_type ON core.domains(type);
CREATE INDEX IF NOT EXISTS idx_domains_whitelist ON core.domains(whitelist);
CREATE INDEX IF NOT EXISTS idx_domain_records_domain_id ON core.domain_records(domain_id);
CREATE INDEX IF NOT EXISTS idx_domain_records_date ON core.domain_records(date);

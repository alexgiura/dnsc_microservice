-- ============================================================================
-- Sports News Ingestion Microservice – Database initialization
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS content;

CREATE TABLE IF NOT EXISTS content.articles (
    id UUID PRIMARY KEY,

    provider TEXT NOT NULL,
    external_id BIGINT NOT NULL,
    type TEXT NOT NULL,

    title TEXT NOT NULL,
    description TEXT,
    summary TEXT,
    body TEXT,

    language TEXT,
    canonical_url TEXT,
    hotlink_url TEXT,
    image_url TEXT,

    published_at TIMESTAMPTZ NOT NULL,
    external_updated_at TIMESTAMPTZ,

    content_hash TEXT NOT NULL,

    sync_status TEXT NOT NULL DEFAULT 'pending',
    sync_attempts INTEGER NOT NULL DEFAULT 0,
    last_synced_at TIMESTAMPTZ,
    sync_error TEXT,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT uq_articles_provider_external_id UNIQUE (provider, external_id)
);

CREATE INDEX IF NOT EXISTS idx_articles_provider
    ON content.articles(provider);

CREATE INDEX IF NOT EXISTS idx_articles_external_id
    ON content.articles(external_id);

CREATE INDEX IF NOT EXISTS idx_articles_type
    ON content.articles(type);

CREATE INDEX IF NOT EXISTS idx_articles_published_at
    ON content.articles(published_at DESC);

CREATE INDEX IF NOT EXISTS idx_articles_external_updated_at
    ON content.articles(external_updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_articles_sync_status
    ON content.articles(sync_status);

CREATE INDEX IF NOT EXISTS idx_articles_last_synced_at
    ON content.articles(last_synced_at);

CREATE INDEX IF NOT EXISTS idx_articles_created_at
    ON content.articles(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_articles_updated_at
    ON content.articles(updated_at DESC);
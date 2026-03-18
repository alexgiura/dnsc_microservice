Backend – Sports News Integration (Go)
=====================================

## Overview

This service periodically fetches sports news articles from an external provider (Pulselive), stores them in PostgreSQL, and synchronizes them with a central system.

The main responsibilities are:
- fetch latest articles from an external API,
- map them to an internal model,
- store them in a database (idempotent upsert),
- send them to a central management system.

---

## Architecture

The project is structured into a few simple layers:

- `internal/external`  
  HTTP clients for external providers (currently Pulselive).

- `internal/models`  
  Data structures and mapping logic:
  - `ExternalArticleDTO` – raw API response  
  - `Article` – internal model  
  - mapping includes a `content_hash` used to detect changes

- `internal/repository`  
  Database layer (PostgreSQL):
  - `Upsert` ensures:
    - insert if new  
    - update if content changed  
    - ignore if unchanged

- `internal/services`  
  Core logic:
  - `ArticleImportService` – fetches and stores articles  
  - `ArticleSyncService` – sends articles to the central system

- `internal/scheduler`  
  Background jobs:
  - import poller (external → DB)  
  - sync poller (DB → central system)

---

## How it works

### 1. Import (external → DB)

A background job runs periodically:
- fetches latest articles from Pulselive (paginated),
- maps them to the internal `Article` model,
- saves them using upsert logic.

Each run processes up to **N = pageSize × maxPages** articles.  
This keeps the system focused on recent content while avoiding large requests.

---

### 2. Sync (DB → central system)

A separate background job:
- selects articles with `sync_status = 'pending'` (and some `failed` with limited attempts),
- sends them via HTTP to the central system,
- marks them as:
  - `synced` (on success),
  - `failed` (on error).

Failed articles are retried in future runs up to a configurable maximum.

---

## Design decisions

- **Idempotency**  
  Articles are re-fetched on every poll.  
  A `content_hash` ensures updates only happen when content actually changes.

- **Separation of import and sync**  
  Import continues even if the central system is down.  
  Sync retries failed articles later, without blocking ingestion.

- **Configurable polling**  
  Intervals, timeouts, pageSize/maxPages and sync retry limits are configurable via environment variables.

- **Retention**  
  Articles are kept in the database.  
  No automatic deletion is implemented in this version.

---

## Running locally

The easiest way is with Docker:

```bash
docker-compose up --build -d
```

This starts:
- PostgreSQL (with the schema from `db/init_scripts`),
- the backend service (built from `backend/Dockerfile` and configured via `.env`).

Once the containers are up:
- health endpoints are available at `http://localhost:8080/health` and `/healthz`,
- import and sync pollers run automatically in the background.

To inspect logs:

```bash
docker-compose logs -f cortex-backend
```

---

## Testing

Basic unit tests are included for:
- article mapping and hashing (`internal/models/article_test.go`),
- import service (`internal/services/article_import_test.go`),
- sync service (`internal/services/article_sync_test.go`).

Run tests from the `backend` directory:

```bash
cd backend
go test ./...
```
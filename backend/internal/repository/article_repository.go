package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cortex/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// UpsertResult represents the outcome of an upsert operation.
type UpsertResult string

const (
	UpsertInserted  UpsertResult = "inserted"
	UpsertUpdated   UpsertResult = "updated"
	UpsertUnchanged UpsertResult = "unchanged"
)

type ArticleRepository interface {
	Upsert(ctx context.Context, article models.Article) (UpsertResult, error)
	GetPendingForSync(ctx context.Context, limit int, maxAttempts int) ([]models.Article, error)
	MarkSynced(ctx context.Context, id uuid.UUID) error
	MarkSyncFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

type articleRepository struct {
	db *pgxpool.Pool
}

func NewArticleRepository(db *pgxpool.Pool) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Upsert(ctx context.Context, article models.Article) (UpsertResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var existingID uuid.UUID
	var existingHash string
	var existingExternalUpdatedAt time.Time

	err = tx.QueryRow(ctx, `
		SELECT id, content_hash, external_updated_at
		FROM content.articles
		WHERE provider = $1 AND external_id = $2
	`, article.Provider, article.ExternalID).Scan(&existingID, &existingHash, &existingExternalUpdatedAt)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("select existing article: %w", err)
	}

	now := time.Now()

	if errors.Is(err, pgx.ErrNoRows) {
		// Insert new article
		_, err = tx.Exec(ctx, `
			INSERT INTO content.articles (
				id, provider, external_id, type,
				title, description, summary, body,
				language, canonical_url, hotlink_url, image_url,
				published_at, external_updated_at,
				content_hash,
				sync_status, sync_attempts, last_synced_at, sync_error,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7, $8,
				$9, $10, $11, $12,
				$13, $14,
				$15,
				$16, $17, $18, $19,
				$20, $21
			)
		`,
			uuid.MustParse(article.ID),
			article.Provider,
			article.ExternalID,
			article.Type,
			article.Title,
			article.Description,
			article.Summary,
			article.Body,
			article.Language,
			article.CanonicalURL,
			article.HotlinkURL,
			article.ImageURL,
			article.PublishedAt,
			article.ExternalUpdatedAt,
			article.ContentHash,
			article.SyncStatus,
			article.SyncAttempts,
			article.LastSyncedAt,
			article.SyncError,
			now,
			now,
		)
		if err != nil {
			return "", fmt.Errorf("insert article: %w", err)
		}
		return UpsertInserted, tx.Commit(ctx)
	} else {
		// Existing article
		if existingHash == article.ContentHash {
			// Content did not change: only update external_updated_at if the external timestamp changed.
			if !article.ExternalUpdatedAt.IsZero() && article.ExternalUpdatedAt.After(existingExternalUpdatedAt) {
				_, err = tx.Exec(ctx, `
					UPDATE content.articles
					SET external_updated_at = $2
					WHERE id = $1
				`, existingID, article.ExternalUpdatedAt)
				if err != nil {
					return "", fmt.Errorf("update article external_updated_at: %w", err)
				}
				return UpsertUnchanged, tx.Commit(ctx)
			}
			// No change to timestamps either
			return UpsertUnchanged, tx.Commit(ctx)
		} else {

			_, err = tx.Exec(ctx, `
				UPDATE content.articles
				SET
					title = $2,
					description = $3,
					summary = $4,
					body = $5,
					language = $6,
					canonical_url = $7,
					hotlink_url = $8,
					image_url = $9,
					published_at = $10,
					external_updated_at = $11,
					content_hash = $12,
					sync_status = 'pending',
					sync_attempts = 0,
					last_synced_at = NULL,
					sync_error = NULL,
					updated_at = $13
				WHERE id = $1
			`,
				existingID,
				article.Title,
				article.Description,
				article.Summary,
				article.Body,
				article.Language,
				article.CanonicalURL,
				article.HotlinkURL,
				article.ImageURL,
				article.PublishedAt,
				article.ExternalUpdatedAt,
				article.ContentHash,
				now,
			)
			if err != nil {
				return "", fmt.Errorf("update article: %w", err)
			}
			return UpsertUpdated, tx.Commit(ctx)
		}
	}
}

func (r *articleRepository) GetPendingForSync(ctx context.Context, limit int, maxAttempts int) ([]models.Article, error) {
	if limit <= 0 {
		limit = 100
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			id, provider, external_id, type,
			title, description, summary, body,
			language, canonical_url, hotlink_url, image_url,
			published_at, external_updated_at,
			content_hash,
			sync_status, sync_attempts, last_synced_at, sync_error,
			created_at, updated_at
		FROM content.articles
		WHERE (sync_status = 'pending' OR (sync_status = 'failed' AND sync_attempts < $2))
		ORDER BY created_at
		LIMIT $1
	`, limit, maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("get pending for sync: %w", err)
	}
	defer rows.Close()

	var result []models.Article
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(
			&a.ID,
			&a.Provider,
			&a.ExternalID,
			&a.Type,
			&a.Title,
			&a.Description,
			&a.Summary,
			&a.Body,
			&a.Language,
			&a.CanonicalURL,
			&a.HotlinkURL,
			&a.ImageURL,
			&a.PublishedAt,
			&a.ExternalUpdatedAt,
			&a.ContentHash,
			&a.SyncStatus,
			&a.SyncAttempts,
			&a.LastSyncedAt,
			&a.SyncError,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pending article: %w", err)
		}
		result = append(result, a)
	}

	return result, rows.Err()
}

func (r *articleRepository) MarkSynced(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE content.articles
		SET sync_status = 'synced',
		    sync_attempts = sync_attempts + 1,
		    last_synced_at = NOW(),
		    sync_error = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("mark synced: %w", err)
	}
	return nil
}

func (r *articleRepository) MarkSyncFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE content.articles
		SET sync_status = 'failed',
		    sync_attempts = sync_attempts + 1,
		    sync_error = $2,
		    updated_at = NOW()
		WHERE id = $1
	`, id, errMsg)
	if err != nil {
		return fmt.Errorf("mark sync failed: %w", err)
	}
	return nil
}

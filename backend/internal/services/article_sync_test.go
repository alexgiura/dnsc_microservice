package services

import (
	"context"
	"errors"
	"testing"

	"cortex/internal/models"
	"cortex/internal/repository"
	"cortex/internal/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCentralClient struct {
	calls   []models.CentralArticleDTO
	failIDs map[string]bool
}

func (c *fakeCentralClient) SyncArticle(ctx context.Context, payload models.CentralArticleDTO) error {
	c.calls = append(c.calls, payload)
	if c.failIDs != nil && c.failIDs[payload.ID] {
		return errors.New("sync failed")
	}
	return nil
}

type fakeSyncRepo struct {
	pending      []models.Article
	markedSynced []uuid.UUID
	markedFailed []uuid.UUID
}

func (r *fakeSyncRepo) GetPendingForSync(ctx context.Context, limit int, maxAttempts int) ([]models.Article, error) {
	return r.pending, nil
}

func (r *fakeSyncRepo) MarkSynced(ctx context.Context, id uuid.UUID) error {
	r.markedSynced = append(r.markedSynced, id)
	return nil
}

func (r *fakeSyncRepo) MarkSyncFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	r.markedFailed = append(r.markedFailed, id)
	return nil
}

// Unused ArticleRepository methods for this fake.
func (r *fakeSyncRepo) Upsert(ctx context.Context, article models.Article) (repository.UpsertResult, error) {
	panic("not implemented")
}

func TestArticleSyncService_MarksSyncedAndFailed(t *testing.T) {
	ctx := context.Background()
	logger := utils.GetLogger("test-sync")

	// Two articles: first will sync ok, second will fail.
	a1 := models.Article{ID: uuid.NewString(), Provider: "pulselive", ExternalID: 1, Title: "A"}
	a2 := models.Article{ID: uuid.NewString(), Provider: "pulselive", ExternalID: 2, Title: "B"}

	repo := &fakeSyncRepo{
		pending: []models.Article{a1, a2},
	}

	client := &fakeCentralClient{
		failIDs: map[string]bool{a2.ID: true},
	}

	svc := NewArticleSyncService(
		repo,
		client,
		logger,
		10, // batch size
		5,  // max attempts
	)

	err := svc.SyncPending(ctx)
	require.NoError(t, err)

	// Should have attempted to sync both.
	assert.Len(t, client.calls, 2)
	assert.Equal(t, a1.ID, client.calls[0].ID)
	assert.Equal(t, a2.ID, client.calls[1].ID)

	// First should be marked synced, second failed.
	require.Len(t, repo.markedSynced, 1)
	require.Len(t, repo.markedFailed, 1)
	assert.Equal(t, a1.ID, repo.markedSynced[0].String())
	assert.Equal(t, a2.ID, repo.markedFailed[0].String())
}

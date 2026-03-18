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

type fakeArticlesClient struct {
	pages [][]models.ExternalArticleDTO
	calls int
}

func (f *fakeArticlesClient) FetchLatest(ctx context.Context, page, pageSize int) ([]models.ExternalArticleDTO, error) {
	if page >= len(f.pages) {
		return []models.ExternalArticleDTO{}, nil
	}
	f.calls++
	return f.pages[page], nil
}

type fakeArticleRepo struct {
	upsertCalls []models.Article
	failOnIndex int // -1 for never
}

func (r *fakeArticleRepo) Upsert(ctx context.Context, article models.Article) (repository.UpsertResult, error) {
	idx := len(r.upsertCalls)
	r.upsertCalls = append(r.upsertCalls, article)
	if r.failOnIndex >= 0 && idx == r.failOnIndex {
		return "", errors.New("upsert failed")
	}
	return repository.UpsertInserted, nil
}

func (r *fakeArticleRepo) GetPendingForSync(ctx context.Context, limit int, maxAttempts int) ([]models.Article, error) {
	panic("not implemented")
}
func (r *fakeArticleRepo) MarkSynced(ctx context.Context, id uuid.UUID) error {
	panic("not implemented")
}
func (r *fakeArticleRepo) MarkSyncFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	panic("not implemented")
}

func TestArticleImportService_ImportsArticles(t *testing.T) {
	ctx := context.Background()

	client := &fakeArticlesClient{
		pages: [][]models.ExternalArticleDTO{
			{
				{ID: 1, Title: "A", Description: "d1", Summary: "s1", Language: "en", CanonicalURL: "c1", HotlinkURL: "h1", ImageURL: "i1", Date: "2024-01-02T15:04:05Z", LastModified: 1704200000000},
				{ID: 2, Title: "B", Description: "d2", Summary: "s2", Language: "en", CanonicalURL: "c2", HotlinkURL: "h2", ImageURL: "i2", Date: "2024-01-02T16:04:05Z", LastModified: 1704203600000},
			},
		},
	}
	repo := &fakeArticleRepo{failOnIndex: -1}
	logger := utils.GetLogger("test-import")

	svc := NewArticleImportService(client, repo, logger, 10, 2)

	err := svc.ImportLatestArticles(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, client.calls, "should fetch exactly one page")
	require.Len(t, repo.upsertCalls, 2, "should upsert both articles")
	assert.Equal(t, int64(1), repo.upsertCalls[0].ExternalID)
	assert.Equal(t, int64(2), repo.upsertCalls[1].ExternalID)
}

func TestArticleImportService_ContinuesOnUpsertError(t *testing.T) {
	ctx := context.Background()

	client := &fakeArticlesClient{
		pages: [][]models.ExternalArticleDTO{
			{
				{ID: 1, Title: "A", Description: "d1", Summary: "s1", Language: "en", CanonicalURL: "c1", HotlinkURL: "h1", ImageURL: "i1", Date: "2024-01-02T15:04:05Z", LastModified: 1704200000000},
				{ID: 2, Title: "B", Description: "d2", Summary: "s2", Language: "en", CanonicalURL: "c2", HotlinkURL: "h2", ImageURL: "i2", Date: "2024-01-02T16:04:05Z", LastModified: 1704203600000},
			},
		},
	}
	repo := &fakeArticleRepo{failOnIndex: 0}
	logger := utils.GetLogger("test-import")

	svc := NewArticleImportService(client, repo, logger, 10, 1)

	err := svc.ImportLatestArticles(ctx)

	require.NoError(t, err)
	assert.Len(t, repo.upsertCalls, 2, "should attempt upsert for both articles even if first fails")
}

package services

import (
	"context"
	"fmt"
	"time"

	"cortex/internal/central"
	"cortex/internal/models"
	"cortex/internal/repository"
	"cortex/internal/utils"

	"github.com/google/uuid"
)

// ArticleSyncService defines operations for synchronizing articles
// with the central management system.
type ArticleSyncService interface {
	SyncPending(ctx context.Context) error
}

type articleSyncService struct {
	repo         repository.ArticleRepository
	client       central.Client
	logger       *utils.Logger
	batchSize    int
	maxAttempts  int
}

// NewArticleSyncService creates a new service for syncing articles.
func NewArticleSyncService(
	repo repository.ArticleRepository,
	client central.Client,
	logger *utils.Logger,
	batchSize int,
	maxAttempts int,
) ArticleSyncService {
	if batchSize <= 0 {
		batchSize = 50
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}

	return &articleSyncService{
		repo:        repo,
		client:      client,
		logger:      logger,
		batchSize:   batchSize,
		maxAttempts: maxAttempts,
	}
}

func (s *articleSyncService) SyncPending(ctx context.Context) error {
	start := time.Now()
	synced := 0
	failed := 0

	articles, err := s.repo.GetPendingForSync(ctx, s.batchSize, s.maxAttempts)
	if err != nil {
		return fmt.Errorf("get pending articles: %w", err)
	}
	fetched := len(articles)

	for _, a := range articles {
		if err := s.syncOne(ctx, a); err != nil {
			failed++
		} else {
			synced++
		}
	}

	duration := time.Since(start).Milliseconds()
	if s.logger != nil {
		s.logger.Info(
			"SYNC",
			"/articles",
			0,
			duration,
			fmt.Sprintf("sync completed: fetched=%d synced=%d failed=%d", fetched, synced, failed),
		)
	}

	return nil
}

func (s *articleSyncService) syncOne(ctx context.Context, article models.Article) error {
	id, err := uuid.Parse(article.ID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("SYNC", "/articles", 0, 0, fmt.Sprintf("invalid article id %q: %v", article.ID, err))
		}
		return err
	}

	payload := models.ArticleToCentralDTO(article)
	if err := s.client.SyncArticle(ctx, payload); err != nil {
		if s.logger != nil {
			s.logger.Error("SYNC", "/articles", 0, 0, fmt.Sprintf("sync article failed (id=%s): %v", article.ID, err))
		}
		if markErr := s.repo.MarkSyncFailed(ctx, id, err.Error()); markErr != nil && s.logger != nil {
			s.logger.Error("SYNC", "/articles", 0, 0, fmt.Sprintf("mark sync failed (id=%s): %v", article.ID, markErr))
		}
		return err
	}

	if err := s.repo.MarkSynced(ctx, id); err != nil {
		if s.logger != nil {
			s.logger.Error("SYNC", "/articles", 0, 0, fmt.Sprintf("mark synced failed (id=%s): %v", article.ID, err))
		}
		return err
	}

	return nil
}


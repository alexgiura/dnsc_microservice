package services

import (
	"context"
	"fmt"
	"time"

	"cortex/internal/external"
	"cortex/internal/models"
	"cortex/internal/repository"
	"cortex/internal/utils"
)

type ArticleImportService interface {
	ImportLatestArticles(ctx context.Context) error
}

type articleImportService struct {
	client       external.ArticlesClient
	articlesRepo repository.ArticleRepository
	logger       *utils.Logger

	pageSize int
	maxPages int
}

func NewArticleImportService(
	client external.ArticlesClient,
	articlesRepo repository.ArticleRepository,
	logger *utils.Logger,
	pageSize int,
	maxPages int,
) ArticleImportService {
	if pageSize <= 0 {
		pageSize = 20
	}
	if maxPages <= 0 {
		maxPages = 1
	}
	return &articleImportService{
		client:       client,
		articlesRepo: articlesRepo,
		logger:       logger,
		pageSize:     pageSize,
		maxPages:     maxPages,
	}
}

func (s *articleImportService) ImportLatestArticles(ctx context.Context) error {
	start := time.Now()
	fetchedCount := 0
	insertedCount := 0
	updatedCount := 0
	unchangedCount := 0
	failedCount := 0
	processedPages := 0

	for page := 0; page < s.maxPages; page++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		dtos, err := s.client.FetchLatest(ctx, page, s.pageSize)
		if err != nil {
			return fmt.Errorf("fetch latest page %d: %w", page, err)
		}
		if len(dtos) == 0 {
			break
		}
		processedPages++
		fetchedCount += len(dtos)

		for _, dto := range dtos {
			article, err := models.MapToArticle(dto)
			if err != nil {
				s.logger.Error("POLL", "/articles", 0, 0, fmt.Sprintf("failed to map external article to internal article (id=%d): %v", dto.ID, err))
				failedCount++
				continue
			}
			result, err := s.articlesRepo.Upsert(ctx, article)
			if err != nil {
				s.logger.Error("POLL", "/articles", 0, 0, fmt.Sprintf("failed to save article (id=%d): %v", dto.ID, err))
				failedCount++
				continue
			}
			switch result {
			case repository.UpsertInserted:
				insertedCount++
			case repository.UpsertUpdated:
				updatedCount++
			case repository.UpsertUnchanged:
				unchangedCount++
			}
		}
	}

	duration := time.Since(start).Milliseconds()
	s.logger.Info(
		"POLL",
		"/articles",
		0,
		duration,
		fmt.Sprintf(
			"article import completed: pages=%d fetched=%d inserted=%d updated=%d unchanged=%d failed=%d",
			processedPages,
			fetchedCount,
			insertedCount,
			updatedCount,
			unchangedCount,
			failedCount,
		),
	)

	return nil
}

package services

import (
	"time"

	"cortex/internal/central"
	"cortex/internal/config"
	"cortex/internal/external"
	"cortex/internal/repository"
	"cortex/internal/scheduler"
	"cortex/internal/utils"
)

type AppServices struct {
	ArticlePoller     *scheduler.ArticlePoller
	ArticleSyncPoller *scheduler.ArticleSyncPoller
}

// NewAppServices initializes all services from config.
func NewAppServices(repos *repository.Repository, logger *utils.Logger, cfg *config.Config) *AppServices {
	if cfg == nil {
		cfg = &config.Config{}
	}

	// External client and article import (from config)
	imp := cfg.ArticleImport
	pulselive := cfg.Pulselive
	pulseliveClient := external.NewPulseliveClient(
		pulselive.BaseURL,
		time.Duration(pulselive.Timeout)*time.Second,
	)

	importService := NewArticleImportService(
		pulseliveClient,
		repos.Article,
		logger,
		imp.PageSize,
		imp.MaxPages,
	)

	importPoller := scheduler.NewArticlePoller(
		importService,
		time.Duration(imp.Interval)*time.Second,
		time.Duration(imp.Timeout)*time.Second,
		logger,
	)

	// Central client and article sync (from config)
	syncCfg := cfg.ArticleSync
	centralCfg := cfg.Central
	centralClient := central.NewHTTPClient(
		centralCfg.BaseURL,
		time.Duration(centralCfg.Timeout)*time.Second,
	)

	syncService := NewArticleSyncService(
		repos.Article,
		centralClient,
		logger,
		syncCfg.BatchSize,
		syncCfg.MaxAttempts,
	)

	syncPoller := scheduler.NewArticleSyncPoller(
		syncService,
		time.Duration(syncCfg.Interval)*time.Second,
		time.Duration(syncCfg.Timeout)*time.Second,
		logger,
	)

	return &AppServices{
		ArticlePoller:     importPoller,
		ArticleSyncPoller: syncPoller,
	}
}

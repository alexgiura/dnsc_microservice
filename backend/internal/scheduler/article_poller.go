package scheduler

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"cortex/internal/utils"
)

type ArticleImportRunner interface {
	ImportLatestArticles(ctx context.Context) error
}

type ArticlePoller struct {
	importService ArticleImportRunner
	interval      time.Duration
	runTimeout    time.Duration
	logger        *utils.Logger
	running       atomic.Bool
}

func NewArticlePoller(
	importService ArticleImportRunner,
	interval time.Duration,
	runTimeout time.Duration,
	logger *utils.Logger,
) *ArticlePoller {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	if runTimeout <= 0 {
		runTimeout = 30 * time.Second
	}
	return &ArticlePoller{
		importService: importService,
		interval:      interval,
		runTimeout:    runTimeout,
		logger:        logger,
	}
}

func (p *ArticlePoller) Start(ctx context.Context) {
	p.runOnce(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("POLL", "/articles", 0, 0, "stopping article poller")
			return
		case <-ticker.C:
			p.runOnce(ctx)
		}
	}
}

func (p *ArticlePoller) runOnce(parentCtx context.Context) {
	if !p.running.CompareAndSwap(false, true) {
		p.logger.Warn("POLL", "/articles", 0, 0, "skipping poll run: already running")
		return
	}
	defer p.running.Store(false)

	ctx, cancel := context.WithTimeout(parentCtx, p.runTimeout)
	defer cancel()

	start := time.Now()
	if err := p.importService.ImportLatestArticles(ctx); err != nil {
		duration := time.Since(start).Milliseconds()
		p.logger.Error("POLL", "/articles", 0, duration, fmt.Sprintf("import latest articles failed: %v", err))
		return
	}

	duration := time.Since(start).Milliseconds()
	p.logger.Info("POLL", "/articles", 0, duration, "import latest articles completed successfully")
}

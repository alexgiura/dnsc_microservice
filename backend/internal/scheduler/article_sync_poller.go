package scheduler

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"cortex/internal/utils"
)

type ArticleSyncRunner interface {
	SyncPending(ctx context.Context) error
}

type ArticleSyncPoller struct {
	syncService ArticleSyncRunner
	interval    time.Duration
	runTimeout  time.Duration
	logger      *utils.Logger

	running atomic.Bool
}

func NewArticleSyncPoller(
	syncService ArticleSyncRunner,
	interval time.Duration,
	runTimeout time.Duration,
	logger *utils.Logger,
) *ArticleSyncPoller {
	if interval <= 0 {
		interval = 1 * time.Minute
	}
	if runTimeout <= 0 {
		runTimeout = 30 * time.Second
	}
	return &ArticleSyncPoller{
		syncService: syncService,
		interval:    interval,
		runTimeout:  runTimeout,
		logger:      logger,
	}
}

// Start runs one sync immediately and then continues on a ticker until ctx is done.
func (p *ArticleSyncPoller) Start(ctx context.Context) {
	p.runOnce(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if p.logger != nil {
				p.logger.Info("SYNC", "/articles", 0, 0, "stopping article sync poller")
			}
			return
		case <-ticker.C:
			p.runOnce(ctx)
		}
	}
}

func (p *ArticleSyncPoller) runOnce(parentCtx context.Context) {
	if !p.running.CompareAndSwap(false, true) {
		if p.logger != nil {
			p.logger.Warn("SYNC", "/articles", 0, 0, "skipping sync run: already running")
		}
		return
	}
	defer p.running.Store(false)

	ctx, cancel := context.WithTimeout(parentCtx, p.runTimeout)
	defer cancel()

	start := time.Now()
	if err := p.syncService.SyncPending(ctx); err != nil {
		duration := time.Since(start).Milliseconds()
		if p.logger != nil {
			p.logger.Error("SYNC", "/articles", 0, duration, fmt.Sprintf("sync pending articles failed: %v", err))
		}
		return
	}

	duration := time.Since(start).Milliseconds()
	if p.logger != nil {
		p.logger.Info("SYNC", "/articles", 0, duration, "sync pending articles completed successfully")
	}
}

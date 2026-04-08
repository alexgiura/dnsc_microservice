package scheduler

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"dnsc_microservice/internal/services"
)

// DomainPNRISCSyncScheduler POSTs changed domains to the PNRISC API on a schedule.
type DomainPNRISCSyncScheduler struct {
	domainSvc services.DomainService
	enabled   bool
	url       string
	schedule  string
	timezone  string

	cron *cron.Cron
}

func NewDomainPNRISCSyncScheduler(
	domainSvc services.DomainService,
	enabled bool,
	url, schedule, timezone string,
) *DomainPNRISCSyncScheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("[pnrisc-sync] invalid timezone=%q: %v; using UTC", timezone, err)
		loc = time.UTC
	}

	parser := cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)

	c := cron.New(
		cron.WithLocation(loc),
		cron.WithParser(parser),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
	)

	return &DomainPNRISCSyncScheduler{
		domainSvc: domainSvc,
		enabled:   enabled,
		url:       strings.TrimSpace(url),
		schedule:  schedule,
		timezone:  timezone,
		cron:      c,
	}
}

func (s *DomainPNRISCSyncScheduler) Start(ctx context.Context) error {
	if !s.enabled {
		return nil
	}
	if s.url == "" {
		log.Printf("[pnrisc-sync] enabled but DOMAIN_PNRISC_SYNC_URL is empty; scheduler not started")
		return nil
	}

	_, err := s.cron.AddFunc(s.schedule, func() {
		runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		log.Printf("[pnrisc-sync] job started; url=%s timezone=%s", s.url, s.timezone)
		if err := s.domainSvc.SyncPNRISCDomains(runCtx); err != nil {
			log.Printf("[pnrisc-sync] job failed: %v", err)
		} else {
			log.Printf("[pnrisc-sync] job completed")
		}
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	log.Printf("[pnrisc-sync] started; schedule=%q timezone=%s", s.schedule, s.timezone)
	return nil
}

func (s *DomainPNRISCSyncScheduler) Stop() {
	if s.cron == nil {
		return
	}
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Printf("[pnrisc-sync] stopped")
}

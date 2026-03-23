package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"dnsc_microservice/internal/services"
)

// DomainAutoWhitelistScheduler runs a periodic job that auto-whitelists domains
// when the latest domain_record is older than N months.
type DomainAutoWhitelistScheduler struct {
	domainSvc      services.DomainService
	enabled        bool
	schedule       string
	timezone       string
	inactivityDays int
	changedBy      string
	notes          string

	cron *cron.Cron
}

func NewDomainAutoWhitelistScheduler(
	domainSvc services.DomainService,
	enabled bool,
	schedule string,
	timezone string,
	inactivityDays int,
	changedBy string,
	notes string,
) *DomainAutoWhitelistScheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("[domain-auto-whitelist] invalid timezone=%q: %v; using UTC", timezone, err)
		loc = time.UTC
	}

	// robfig/cron uses seconds as the first field when configured with the Seconds parser.
	parser := cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)

	c := cron.New(
		cron.WithLocation(loc),
		cron.WithParser(parser),
		cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)),
	)

	return &DomainAutoWhitelistScheduler{
		domainSvc:      domainSvc,
		enabled:        enabled,
		schedule:       schedule,
		timezone:       timezone,
		inactivityDays: inactivityDays,
		changedBy:      changedBy,
		notes:          notes,
		cron:           c,
	}
}

func (s *DomainAutoWhitelistScheduler) Start(ctx context.Context) error {
	if !s.enabled {
		return nil
	}

	// Each run should have its own timeout so a slow DB doesn’t stall the scheduler.
	_, err := s.cron.AddFunc(s.schedule, func() {
		runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Compute cutoff based on inactivity window (in days).
		cutoff := time.Now().Add(-time.Hour * 24 * time.Duration(s.inactivityDays))

		log.Printf("[domain-auto-whitelist] job started; cutoff=%s timezone=%s", cutoff.Format(time.RFC3339), s.timezone)
		if err := s.domainSvc.AutoWhitelistStaleDomains(runCtx, cutoff, s.changedBy, s.notes); err != nil {
			log.Printf("[domain-auto-whitelist] job failed: %v", err)
		} else {
			log.Printf("[domain-auto-whitelist] job completed")
		}
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	log.Printf("[domain-auto-whitelist] started; schedule=%q timezone=%s", s.schedule, s.timezone)
	return nil
}

func (s *DomainAutoWhitelistScheduler) Stop() {
	if s.cron == nil {
		return
	}
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Printf("[domain-auto-whitelist] stopped")
}

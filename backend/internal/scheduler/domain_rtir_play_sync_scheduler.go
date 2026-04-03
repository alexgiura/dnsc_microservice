package scheduler

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"dnsc_microservice/internal/services"
)

// DomainRTIRPlaySyncScheduler periodically GETs RTIR /REST/2.0/tickets. URL, token, schedule, timezone come from config.
type DomainRTIRPlaySyncScheduler struct {
	domainSvc  services.DomainService
	enabled    bool
	url        string
	schedule   string
	timezone   string
	token string

	cron *cron.Cron
}

func NewDomainRTIRPlaySyncScheduler(
	domainSvc services.DomainService,
	enabled bool,
	url, schedule, timezone, token string,
) *DomainRTIRPlaySyncScheduler {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("[rtir-play-sync] invalid timezone=%q: %v; using UTC", timezone, err)
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

	return &DomainRTIRPlaySyncScheduler{
		domainSvc:  domainSvc,
		enabled:    enabled,
		url:        strings.TrimSpace(url),
		schedule:   schedule,
		timezone:   timezone,
		token: strings.TrimSpace(token),
		cron:  c,
	}
}

func (s *DomainRTIRPlaySyncScheduler) Start(ctx context.Context) error {
	if !s.enabled {
		return nil
	}
	if s.url == "" {
		log.Printf("[rtir-play-sync] enabled but DOMAIN_RTIR_PLAY_SYNC_URL is empty; scheduler not started")
		return nil
	}
	if s.token == "" {
		log.Printf("[rtir-play-sync] enabled but DOMAIN_RTIR_PLAY_SYNC_TOKEN is empty; scheduler not started")
		return nil
	}

	_, err := s.cron.AddFunc(s.schedule, func() {
		runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		log.Printf("[rtir-play-sync] job started; url=%s timezone=%s", s.url, s.timezone)
		if err := s.domainSvc.SyncRTIRPlayDomains(runCtx); err != nil {
			log.Printf("[rtir-play-sync] job failed: %v", err)
		} else {
			log.Printf("[rtir-play-sync] job completed")
		}
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	log.Printf("[rtir-play-sync] started; schedule=%q timezone=%s", s.schedule, s.timezone)
	return nil
}

func (s *DomainRTIRPlaySyncScheduler) Stop() {
	if s.cron == nil {
		return
	}
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Printf("[rtir-play-sync] stopped")
}

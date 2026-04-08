package app

import (
	"context"
	"fmt"
	"log"
	"strings"

	"dnsc_microservice/internal/clients/pnrisc"
	"dnsc_microservice/internal/clients/rtir"
	"dnsc_microservice/internal/config"
	"dnsc_microservice/internal/db"
	"dnsc_microservice/internal/middleware"
	"dnsc_microservice/internal/repository"
	"dnsc_microservice/internal/routes"
	"dnsc_microservice/internal/scheduler"
	"dnsc_microservice/internal/server"
	"dnsc_microservice/internal/services"

	"github.com/jackc/pgx/v4/pgxpool"
)

type App struct {
	server        *server.Server
	dbPool        *pgxpool.Pool
	sched         *scheduler.DomainAutoWhitelistScheduler
	rtirPlaySched *scheduler.DomainRTIRPlaySyncScheduler
	pnriscSched   *scheduler.DomainPNRISCSyncScheduler
}

func NewApp(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	pool, err := db.NewPostgresPool(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	repo := repository.NewRepository(pool)

	rtirClient := rtir.NewClient(rtir.Config{
		BaseURL:       cfg.DomainRTIRPlaySyncSettings.URL,
		Token:         cfg.DomainRTIRPlaySyncSettings.Token,
		SkipTLSVerify: cfg.DomainRTIRPlaySyncSettings.SkipTLSVerify,
	})

	var pnriscClient *pnrisc.Client
	if strings.TrimSpace(cfg.DomainPNRISCSyncSettings.URL) != "" {
		pnriscClient = pnrisc.NewClient(pnrisc.Config{
			BaseURL:       cfg.DomainPNRISCSyncSettings.URL,
			Token:         cfg.DomainPNRISCSyncSettings.Token,
			SkipTLSVerify: cfg.DomainPNRISCSyncSettings.SkipTLSVerify,
		})
	}

	appServices := services.NewAppServices(
		repo,
		rtirClient,
		pnriscClient,
		cfg.DomainRTIRPlaySyncSettings.Timezone,
		cfg.DomainRTIRPlaySyncSettings.OverlapMinutes,
	)

	autoScheduler := scheduler.NewDomainAutoWhitelistScheduler(
		appServices.Domain,
		cfg.DomainAutoWhitelistSettings.Enabled,
		cfg.DomainAutoWhitelistSettings.Schedule,
		cfg.DomainAutoWhitelistSettings.Timezone,
		cfg.DomainAutoWhitelistSettings.InactivityDays,
		cfg.DomainAutoWhitelistSettings.ChangedBy,
		cfg.DomainAutoWhitelistSettings.Notes,
	)

	rtirPlayScheduler := scheduler.NewDomainRTIRPlaySyncScheduler(
		appServices.Domain,
		cfg.DomainRTIRPlaySyncSettings.Enabled,
		cfg.DomainRTIRPlaySyncSettings.URL,
		cfg.DomainRTIRPlaySyncSettings.Schedule,
		cfg.DomainRTIRPlaySyncSettings.Timezone,
		cfg.DomainRTIRPlaySyncSettings.Token,
	)

	pnriscScheduler := scheduler.NewDomainPNRISCSyncScheduler(
		appServices.Domain,
		cfg.DomainPNRISCSyncSettings.Enabled,
		cfg.DomainPNRISCSyncSettings.URL,
		cfg.DomainPNRISCSyncSettings.Schedule,
		cfg.DomainPNRISCSyncSettings.Timezone,
	)

	router := routes.RegisterRoutes(appServices)

	handlerWithMiddleware := middleware.CorsMiddleware(router)

	srv, err := server.NewServer(cfg.AppSettings.ServerPort, handlerWithMiddleware)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("create server: %w", err)
	}

	return &App{
		server:        srv,
		dbPool:        pool,
		sched:         autoScheduler,
		rtirPlaySched: rtirPlayScheduler,
		pnriscSched:   pnriscScheduler,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	// _ = ctx
	if app.server == nil {
		return fmt.Errorf("server is nil")
	}

	if app.sched != nil {
		if err := app.sched.Start(ctx); err != nil {
			return fmt.Errorf("start auto-whitelist scheduler: %w", err)
		}
	}
	if app.rtirPlaySched != nil {
		if err := app.rtirPlaySched.Start(ctx); err != nil {
			return fmt.Errorf("start RTIR Play sync scheduler: %w", err)
		}
	}
	if app.pnriscSched != nil {
		if err := app.pnriscSched.Start(ctx); err != nil {
			return fmt.Errorf("start PNRISC sync scheduler: %w", err)
		}
	}

	log.Println("starting HTTP server")

	if err := app.server.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func (app *App) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("shutdown context is nil")
	}

	log.Println("shutting down application")

	var shutdownErr error

	if app.server != nil {
		if err := app.server.Shutdown(ctx); err != nil {
			log.Printf("error shutting down server: %v", err)
			if shutdownErr == nil {
				shutdownErr = fmt.Errorf("shutdown server: %w", err)
			}
		} else {
			log.Println("server stopped successfully")
		}
	}

	if app.sched != nil {
		app.sched.Stop()
	}
	if app.rtirPlaySched != nil {
		app.rtirPlaySched.Stop()
	}
	if app.pnriscSched != nil {
		app.pnriscSched.Stop()
	}

	// Close PostgreSQL connection pool
	if app.dbPool != nil {
		app.dbPool.Close()
		log.Println("postgres connection pool closed")
	}

	return shutdownErr
}

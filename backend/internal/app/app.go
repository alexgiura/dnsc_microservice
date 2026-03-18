package app

import (
	"context"
	"fmt"
	"log"

	"cortex/internal/config"
	"cortex/internal/db"
	"cortex/internal/middleware"
	"cortex/internal/repository"
	"cortex/internal/routes"
	"cortex/internal/scheduler"
	"cortex/internal/server"
	"cortex/internal/services"
	"cortex/internal/utils"

	"github.com/jackc/pgx/v4/pgxpool"
)

type App struct {
	server *server.Server
	dbPool *pgxpool.Pool
	logger *utils.Logger

	articlePoller     *scheduler.ArticlePoller
	articleSyncPoller *scheduler.ArticleSyncPoller
}

func NewApp(cfg *config.Config, logger *utils.Logger) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	pool, err := db.NewPostgresPool(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	repo := repository.NewRepository(pool)

	appServices := services.NewAppServices(repo, logger, cfg)

	router := routes.RegisterRoutes()

	handlerWithMiddleware := middleware.CorsMiddleware(router)

	srv, err := server.NewServer(cfg.AppSettings.ServerPort, handlerWithMiddleware)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("create server: %w", err)
	}

	return &App{
		server:            srv,
		dbPool:            pool,
		logger:            logger,
		articlePoller:     appServices.ArticlePoller,
		articleSyncPoller: appServices.ArticleSyncPoller,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	if app.server == nil {
		return fmt.Errorf("server is nil")
	}

	// Start background article poller
	if app.articlePoller != nil {
		go app.articlePoller.Start(ctx)
	}

	// Start background article sync poller
	if app.articleSyncPoller != nil {
		go app.articleSyncPoller.Start(ctx)
	}

	if app.logger != nil {
		app.logger.Info("APP", "/startup", 0, 0, "starting HTTP server")
	} else {
		log.Println("starting HTTP server")
	}

	if err := app.server.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}

func (app *App) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("shutdown context is nil")
	}

	if app.logger != nil {
		app.logger.Info("APP", "/shutdown", 0, 0, "shutting down application")
	} else {
		log.Println("shutting down application")
	}

	var shutdownErr error

	if app.server != nil {
		if err := app.server.Shutdown(ctx); err != nil {
			if app.logger != nil {
				app.logger.Error("APP", "/shutdown", 0, 0, fmt.Sprintf("error shutting down server: %v", err))
			} else {
				log.Printf("error shutting down server: %v", err)
			}
			if shutdownErr == nil {
				shutdownErr = fmt.Errorf("shutdown server: %w", err)
			}
		} else {
			if app.logger != nil {
				app.logger.Info("APP", "/shutdown", 0, 0, "server stopped successfully")
			} else {
				log.Println("server stopped successfully")
			}
		}
	}

	// Close PostgreSQL connection pool
	if app.dbPool != nil {
		app.dbPool.Close()
		if app.logger != nil {
			app.logger.Info("APP", "/shutdown", 0, 0, "postgres connection pool closed")
		} else {
			log.Println("postgres connection pool closed")
		}
	}

	return shutdownErr
}

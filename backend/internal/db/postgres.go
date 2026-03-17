package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"dnsc_microservice/internal/config"

	"github.com/jackc/pgx/v4/pgxpool"
)

// NewPostgresPool creates and verifies a PostgreSQL connection pool using the provided config.
func NewPostgresPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	dbConnectString := cfg.PostgreSQLConnectionString()

	poolConfig, err := pgxpool.ParseConfig(dbConnectString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Connection pool settings
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	const maxRetries = 3
	const retryDelay = 1 * time.Second

	var pool *pgxpool.Pool
	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.ConnectConfig(ctx, poolConfig)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				log.Println("✅ Successfully connected to PostgreSQL")
				return pool, nil
			} else {
				err = pingErr
			}
		}

		log.Printf("Retrying database connection (%d/%d): %v\n", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("unable to establish database connection after %d retries: %w", maxRetries, err)
}


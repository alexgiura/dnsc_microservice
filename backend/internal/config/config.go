package config

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

type AppSettings struct {
	ServerPort  string `env:"SERVER_PORT" envDefault:"8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	DebugMode   bool   `env:"DEBUG_MODE" envDefault:"false"`
}

type DatabaseSettings struct {
	User     string `env:"POSTGRES_DB_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_DB_PASSWORD" envDefault:"postgres"`
	Host     string `env:"POSTGRES_DB_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_DB_PORT" envDefault:"5432"`
	DbName   string `env:"POSTGRES_DB_NAME" envDefault:"dnsc_db"`
	SSLMode  string `env:"POSTGRES_DB_SSLMODE" envDefault:"require"`
}

type PulseliveConfig struct {
	BaseURL string `env:"PULSELIVE_BASE_URL" envDefault:"https://content-ecb.pulselive.com/content/ecb/text/EN/"`
	Timeout int    `env:"PULSELIVE_TIMEOUT" envDefault:"5"`
}

type CentralConfig struct {
	BaseURL string `env:"CENTRAL_BASE_URL" envDefault:"http://central-management-system:8080"`
	Timeout int    `env:"CENTRAL_TIMEOUT" envDefault:"5"`
}

type ArticleImportConfig struct {
	PageSize int `env:"ARTICLE_PAGE_SIZE" envDefault:"20"`
	MaxPages int `env:"ARTICLE_MAX_PAGES" envDefault:"3"`
	Interval int `env:"ARTICLE_IMPORT_INTERVAL" envDefault:"60"`
	Timeout  int `env:"ARTICLE_IMPORT_TIMEOUT" envDefault:"30"`
}

type ArticleSyncConfig struct {
	BatchSize   int `env:"ARTICLE_SYNC_BATCH_SIZE" envDefault:"50"`
	MaxAttempts int `env:"ARTICLE_SYNC_MAX_ATTEMPTS" envDefault:"5"`
	Interval    int `env:"ARTICLE_SYNC_INTERVAL" envDefault:"120"`
	Timeout     int `env:"ARTICLE_SYNC_TIMEOUT" envDefault:"30"`
}

type Config struct {
	AppSettings      AppSettings
	DatabaseSettings DatabaseSettings
	Pulselive        PulseliveConfig
	Central          CentralConfig
	ArticleImport    ArticleImportConfig
	ArticleSync      ArticleSyncConfig
}

func ConnectPostgreSQL(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	dbConnectString := cfg.PostgreSQLConnectionString()

	config, err := pgxpool.ParseConfig(dbConnectString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	const maxRetries = 3
	const retryDelay = 1 * time.Second

	var pool *pgxpool.Pool
	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.ConnectConfig(ctx, config)
		if err == nil {
			// Perform a health check
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

func Load() (*Config, error) {
	cfg := &Config{}

	// Get the current file path
	_, currentFilePath, _, _ := runtime.Caller(0)

	// Navigate to backend directory:
	backendPath := filepath.Join(filepath.Dir(currentFilePath), "..", "..", "..")
	envFilePath := filepath.Join(backendPath, ".env")

	// Load the .env file from backend directory
	err := godotenv.Load(envFilePath)
	if err != nil {
		log.Printf("No .env file found at %s, using environment variables.\n", envFilePath)
	} else {
		log.Printf("✅ Loaded .env file from: %s\n", envFilePath)
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("error loading configuration: %s", err)
	}

	if cfg.AppSettings.ServerPort == "" {
		return nil, fmt.Errorf("invalid config: SERVER_PORT must not be empty")
	}

	log.Printf("✅ Loaded config - ServerPort: %s, Environment: %s\n",
		cfg.AppSettings.ServerPort, cfg.AppSettings.Environment)

	return cfg, nil
}

func (cfg *Config) PostgreSQLConnectionString() string {
	sslMode := cfg.DatabaseSettings.SSLMode
	if sslMode == "" {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DatabaseSettings.User,
		cfg.DatabaseSettings.Password,
		cfg.DatabaseSettings.Host,
		cfg.DatabaseSettings.Port,
		cfg.DatabaseSettings.DbName,
		sslMode,
	)
}

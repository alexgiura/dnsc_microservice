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

// AppSettings holds configuration for the API server.
type AppSettings struct {
	ServerPort  string `env:"SERVER_PORT" envDefault:"8080"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	DebugMode   bool   `env:"DEBUG_MODE" envDefault:"false"`
}

type DomainAutoWhitelistSettings struct {
	Enabled  bool   `env:"DOMAIN_AUTO_WHITELIST_ENABLED" envDefault:"false"`
	Schedule string `env:"DOMAIN_AUTO_WHITELIST_SCHEDULE" envDefault:"0 0 2 * * *"` // seconds min hour dom mon dow
	Timezone string `env:"DOMAIN_AUTO_WHITELIST_TIMEZONE" envDefault:"UTC"`
	// Inactivity window used to compute cutoff from core.domain_records.date.
	InactivityDays int `env:"DOMAIN_INACTIVITY_DAYS" envDefault:"180"`
	ChangedBy      string `env:"DOMAIN_AUTO_WHITELIST_CHANGED_BY" envDefault:"system"`
	Notes          string `env:"DOMAIN_AUTO_WHITELIST_NOTES" envDefault:"Auto-whitelisted by system."`
}

// DatabaseSettings holds configuration related to the PostgreSQL database.
type DatabaseSettings struct {
	User     string `env:"POSTGRES_DB_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_DB_PASSWORD" envDefault:"postgres"`
	Host     string `env:"POSTGRES_DB_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_DB_PORT" envDefault:"5432"`
	DbName   string `env:"POSTGRES_DB_NAME" envDefault:"dnsc_db"`
	SSLMode  string `env:"POSTGRES_DB_SSLMODE" envDefault:"require"`
}

// Config holds configuration for the API and database.
type Config struct {
	AppSettings      AppSettings
	DatabaseSettings DatabaseSettings
	DomainAutoWhitelistSettings DomainAutoWhitelistSettings
}

// ConnectPostgreSQL connects to PostgreSQL database and returns a connection pool
func ConnectPostgreSQL(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	dbConnectString := cfg.PostgreSQLConnectionString()
	
	config, err := pgxpool.ParseConfig(dbConnectString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Configure connection pool settings
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	// Retry logic for establishing the connection pool
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

// Load loads the configuration from environment variables and returns a Config struct.
func Load() (*Config, error) {
	cfg := &Config{}

	// Load environment variables from .env file in backend directory
	// Get the current file path (this file: dnsc_microservice/internal/config/config.go)
	_, currentFilePath, _, _ := runtime.Caller(0)
	
	// Navigate to backend directory: 
	// dnsc_microservice/internal/config -> dnsc_microservice/internal -> dnsc_microservice -> backend
	backendPath := filepath.Join(filepath.Dir(currentFilePath), "..", "..", "..")
	envFilePath := filepath.Join(backendPath, ".env")
	
	// Load the .env file from backend directory
	err := godotenv.Load(envFilePath)
	if err != nil {
		log.Printf("No .env file found at %s, using environment variables.\n", envFilePath)
	} else {
		log.Printf("✅ Loaded .env file from: %s\n", envFilePath)
	}

	// Parse the configuration from environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("error loading configuration: %s", err)
	}

	// Basic validation for required app settings
	if cfg.AppSettings.ServerPort == "" {
		return nil, fmt.Errorf("invalid config: SERVER_PORT must not be empty")
	}
	
	log.Printf("✅ Loaded config - ServerPort: %s, Environment: %s\n",
		cfg.AppSettings.ServerPort, cfg.AppSettings.Environment)

	return cfg, nil
}

// PostgreSQLConnectionString generates the PostgreSQL connection string
func (cfg *Config) PostgreSQLConnectionString() string {
	sslMode := cfg.DatabaseSettings.SSLMode
	if sslMode == "" {
		sslMode = "require" // Default to require SSL for security
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

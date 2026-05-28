package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool global PostgreSQL connection pool
var Pool *pgxpool.Pool

// ConnectPostgres initializes connection pool to PostgreSQL using DATABASE_URL env
func ConnectPostgres() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("unable to parse DATABASE_URL: %w", err)
	}

	// Connection Pool optimization for high performance production VPS
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 15 * time.Minute

	// Establish connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Ping database to verify connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Pool = pool
	println("[Postgres] Successfully connected to PostgreSQL Database Pool.")
	return nil
}

// ClosePostgres safely closes connection pool
func ClosePostgres() {
	if Pool != nil {
		Pool.Close()
		println("[Postgres] Database Pool closed.")
	}
}

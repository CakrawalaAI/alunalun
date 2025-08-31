package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/radjathaher/alunalun/api/internal/config"
	"github.com/radjathaher/alunalun/api/internal/repository"
)

type Config struct {
	URL               string
	MaxConnections    int32
	MinConnections    int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type DB struct {
	pool    *pgxpool.Pool
	queries *repository.Queries
	logger  *zap.SugaredLogger
}

func New(dbConfig config.DBConfig, logger *zap.SugaredLogger) (*DB, error) {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&statement_timeout=30s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		dbConfig.SSLMode,
	)

	if dbURL == "" {
		logger.Error("Database URL cannot be empty")
		return nil, fmt.Errorf("database URL cannot be empty")
	}

	cfg := DefaultConfig()
	cfg.URL = dbURL

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		logger.Errorf("Failed to parse database URL: %v", err)
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	if cfg.MaxConnections == 0 {
		cfg.MaxConnections = 10
	}
	if cfg.MinConnections == 0 {
		cfg.MinConnections = 2
	}
	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = 30 * time.Minute
	}
	if cfg.MaxConnIdleTime == 0 {
		cfg.MaxConnIdleTime = 15 * time.Minute
	}
	if cfg.HealthCheckPeriod == 0 {
		cfg.HealthCheckPeriod = 5 * time.Minute
	}

	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logger.Errorf("Failed to create connection pool: %v", err)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		logger.Errorf("Failed to ping database: %v", err)
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := repository.New(pool)

	return &DB{
		pool:    pool,
		queries: queries,
		logger:  logger,
	}, nil
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) Queries() *repository.Queries {
	return db.queries
}

func (db *DB) Health(ctx context.Context) error {
	if err := db.pool.Ping(ctx); err != nil {
		db.logger.Errorf("Database health check failed: %v", err)
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		db.logger.Errorf("Failed to start transaction: %v", err)
		return nil, err
	}
	return tx, nil
}

func (db *DB) WithTx(tx pgx.Tx) *repository.Queries {
	return repository.New(tx)
}

func (db *DB) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return db.pool.Acquire(ctx)
}

func (db *DB) Stats() *pgxpool.Stat {
	return db.pool.Stat()
}

func DefaultConfig() Config {
	return Config{
		MaxConnections:    10,
		MinConnections:    2,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   15 * time.Minute,
		HealthCheckPeriod: 5 * time.Minute,
	}
}

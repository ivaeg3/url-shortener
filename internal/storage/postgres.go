package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ivaeg3/url-shortener/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns = 20
	defaultMinConns = 5
	connTimeout     = 5 * time.Second
	queryTimeout    = 3 * time.Second
)

type PostgresStorage struct {
	pool    *pgxpool.Pool
	log     *slog.Logger
	counter uint64
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = defaultMaxConns
	config.MinConns = defaultMinConns
	config.ConnConfig.ConnectTimeout = connTimeout

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log := slog.Default()
	if err := runMigrations(pool, log); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	return &PostgresStorage{
		pool: pool,
		log:  log,
	}, nil
}

func runMigrations(pool *pgxpool.Pool, log *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	_, err := pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS urls (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            original_url TEXT NOT NULL UNIQUE,
            short_url VARCHAR(10) NOT NULL UNIQUE
        );

        CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
        CREATE UNIQUE INDEX IF NOT EXISTS idx_short_url ON urls(short_url);
    `)
	if err != nil {
		log.Error("failed to run migrations", "error", err)
	}
	return err
}

func (s *PostgresStorage) Save(originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var shortURL string

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		s.log.Error("failed to begin transaction", "error", err)
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, "SELECT short_url FROM urls WHERE original_url = $1", originalURL).Scan(&shortURL)
	if err == nil {
		_ = tx.Commit(ctx)
		return shortURL, nil
	}

	if err != pgx.ErrNoRows {
		s.log.Error("failed to check existing URL", "error", err, "originalURL", originalURL)
		return "", fmt.Errorf("failed to check existing URL: %w", err)
	}

	shortURL = utils.Encode(s.counter)

	_, err = tx.Exec(ctx, `
        INSERT INTO urls (original_url, short_url)
        VALUES ($1, $2)
        ON CONFLICT (original_url) DO NOTHING`,
		originalURL, shortURL,
	)
	if err != nil {
		s.log.Error("insert failed", "error", err, "originalURL", originalURL, "shortURL", shortURL)
		return "", fmt.Errorf("insert failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("commit failed", "error", err)
		return "", fmt.Errorf("commit failed: %w", err)
	}

	s.counter++
	return shortURL, nil
}

func (s *PostgresStorage) Get(shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	var originalURL string
	err := s.pool.QueryRow(ctx, `
        SELECT original_url FROM urls 
        WHERE short_url = $1`,
		shortURL,
	).Scan(&originalURL)

	if err == pgx.ErrNoRows {
		return "", ErrNotFound
	}

	if err != nil {
		s.log.Error("failed to get original URL", "error", err, "shortURL", shortURL)
		return "", fmt.Errorf("query failed: %w", err)
	}

	return originalURL, nil
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}

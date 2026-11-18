package postgres

import (
	"context"
	"errors"
	"fmt"

	"url-shortener/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) SaveURL(ctx context.Context, shortCode, originalURL string) error {
	query := `INSERT INTO urls (short_code, original_url) VALUES ($1, $2)`
	_, err := s.pool.Exec(ctx, query, shortCode, originalURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrURLExists
		}
		return fmt.Errorf("insert url: %w", err)
	}
	return nil
}

func (s *Storage) GetURL(ctx context.Context, shortCode string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_code = $1`
	var originalURL string
	err := s.pool.QueryRow(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrURLNotFound
		}
		return "", fmt.Errorf("select url: %w", err)
	}
	return originalURL, nil
}

package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides access to Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// New creates a new repository using DSN from POSTGRES_DSN.
func New(ctx context.Context) (*Repository, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("POSTGRES_DSN is not set")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 4
	cfg.AcquireTimeout = 5 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Repository{pool: pool}, nil
}

// Close closes underlying pool.
func (r *Repository) Close() {
	r.pool.Close()
}

// Result represents a saved bot result.
type Result struct {
	ID        int64
	ChatID    int64
	Data      string
	CreatedAt time.Time
}

// SaveResult inserts bot result and returns its ID.
func (r *Repository) SaveResult(ctx context.Context, chatID int64, data string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `INSERT INTO bot_results (chat_id, data) VALUES ($1, $2) RETURNING id`, chatID, data).Scan(&id)
	return id, err
}

// GetResult retrieves result by ID.
func (r *Repository) GetResult(ctx context.Context, id int64) (*Result, error) {
	var res Result
	err := r.pool.QueryRow(ctx, `SELECT id, chat_id, data, created_at FROM bot_results WHERE id=$1`, id).Scan(
		&res.ID, &res.ChatID, &res.Data, &res.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

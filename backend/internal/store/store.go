package store

import (
	"context"
	"fmt"

	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a Postgres pool and sqlc-generated queries.
type Store struct {
	pool *pgxpool.Pool
	Q    *gen.Queries
}

// New creates a Store from a database URL.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}
	return &Store{
		pool: pool,
		Q:    gen.New(pool),
	}, nil
}

// NewFromPool wraps an existing pool (tests).
func NewFromPool(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, Q: gen.New(pool)}
}

// Pool exposes the underlying connection pool.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Close closes the pool.
func (s *Store) Close() {
	s.pool.Close()
}

// WithTx runs fn inside a transaction. The provided Queries use the tx.
func (s *Store) WithTx(ctx context.Context, fn func(q *gen.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.Q.WithTx(tx)
	if err := fn(q); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// WithTxOptions runs fn inside a transaction with custom pgx options.
func (s *Store) WithTxOptions(ctx context.Context, opts pgx.TxOptions, fn func(q *gen.Queries) error) error {
	tx, err := s.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.Q.WithTx(tx)
	if err := fn(q); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

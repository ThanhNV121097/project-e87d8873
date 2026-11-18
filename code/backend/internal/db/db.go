// Package db opens the Postgres connection pool.
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connectTimeout is how long Open waits for a database that is still
// starting: the runtime creates the database and the backend at the same time.
const connectTimeout = 30 * time.Second

// Open returns a pool that has answered one ping, retrying once a second
// until connectTimeout or until ctx is cancelled.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		// The URL holds the password, so it is never echoed back.
		return nil, errors.New("DATABASE_URL is not a valid connection string")
	}
	cfg.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create the connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	for {
		err = pool.Ping(ctx)
		if err == nil {
			return pool, nil
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("the database did not answer within %s: %w", connectTimeout, err)
		case <-time.After(time.Second):
		}
	}
}

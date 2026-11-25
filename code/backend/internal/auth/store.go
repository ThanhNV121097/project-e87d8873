package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// DB is the part of the connection pool this package uses. server.DB has the
// same method, so the pool satisfies both.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Sessions is where sessions live. Hashes only: the token itself is never
// stored.
type Sessions interface {
	// Create stores a session and purges the expired ones, in one statement.
	Create(ctx context.Context, hash []byte, email string, now, expires time.Time) error
	// Exists reports whether a session with this hash, for this email, has
	// not expired at now.
	Exists(ctx context.Context, hash []byte, email string, now time.Time) (bool, error)
	// Delete removes a session; none found is not an error.
	Delete(ctx context.Context, hash []byte) error
}

// The three statements of the sessions table. One clock, the application's,
// decides every expiry: $3 and $4 are computed in Go, never now().
const (
	// sqlCreate purges what has expired and stores the new session.
	sqlCreate = `WITH purged AS (DELETE FROM "sessions" WHERE "expires_at" <= $3::timestamptz RETURNING 1)
INSERT INTO "sessions" ("token_hash", "email", "expires_at") VALUES ($1::bytea, $2::text, $4::timestamptz)
RETURNING "id"::text`

	// sqlExists is the me check and the guard on every write: one row or none.
	sqlExists = `SELECT 1 FROM "sessions" WHERE "token_hash" = $1::bytea AND "email" = $2::text AND "expires_at" > $3::timestamptz`

	// sqlDelete is sign-out; no row is still success.
	sqlDelete = `DELETE FROM "sessions" WHERE "token_hash" = $1::bytea RETURNING 1`
)

type pgSessions struct{ db DB }

func (s pgSessions) Create(ctx context.Context, hash []byte, email string, now, expires time.Time) error {
	var id string
	return s.db.QueryRow(ctx, sqlCreate, hash, email, now, expires).Scan(&id)
}

func (s pgSessions) Exists(ctx context.Context, hash []byte, email string, now time.Time) (bool, error) {
	var one int
	err := s.db.QueryRow(ctx, sqlExists, hash, email, now).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s pgSessions) Delete(ctx context.Context, hash []byte) error {
	var one int
	err := s.db.QueryRow(ctx, sqlDelete, hash).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

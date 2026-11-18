// Package migrate applies the SQL files in code/backend/migrations to the
// database when the server starts.
//
// The runtime creates an empty database and nothing else applies a schema, so
// the server migrates itself before it listens. Applied versions are recorded
// in schema_migrations, which makes running it again a no-op.
package migrate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// lockKey names this application's migrations to pg_advisory_lock, so two
// processes starting together apply each file once, one after the other.
const lockKey int64 = 7351204118

const (
	ensureTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version    text PRIMARY KEY,
	applied_at timestamptz NOT NULL DEFAULT now()
)`
	selectApplied = `SELECT version FROM schema_migrations`
	insertApplied = `INSERT INTO schema_migrations (version) VALUES ($1)`
)

// Migration is one .up.sql file. Version is its name without the suffix, for
// example 20261118000000_baseline.
type Migration struct {
	Version string
	SQL     string
}

// Load reads every *.up.sql file in the root of fsys, in filename order. The
// .down.sql files are for a person to run by hand and are never applied.
func Load(fsys fs.FS) ([]Migration, error) {
	names, err := fs.Glob(fsys, "*.up.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	out := make([]Migration, 0, len(names))
	for _, name := range names {
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		out = append(out, Migration{Version: strings.TrimSuffix(name, ".up.sql"), SQL: string(body)})
	}
	return out, nil
}

// Up applies every migration in fsys that schema_migrations does not list.
//
// A file runs in one transaction together with its schema_migrations row, so a
// failure leaves the database as it was before that file. A file that holds
// CREATE INDEX CONCURRENTLY cannot run in a transaction: it runs one statement
// at a time and is recorded last, so keep such a file to that one idempotent
// statement (CREATE INDEX CONCURRENTLY IF NOT EXISTS …).
func Up(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS) error {
	all, err := Load(fsys)
	if err != nil {
		return err
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire a connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		return fmt.Errorf("take the migration lock: %w", err)
	}
	defer unlock(conn)

	if _, err := conn.Exec(ctx, ensureTable); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}
	for _, m := range all {
		if applied[m.Version] {
			continue
		}
		if err := apply(ctx, conn, m); err != nil {
			return fmt.Errorf("migration %s: %w", m.Version, err)
		}
		slog.Info("migration applied", "version", m.Version)
	}
	return nil
}

// unlock releases the advisory lock. A connection that may still hold it is
// closed rather than returned to the pool, which frees the lock.
func unlock(conn *pgxpool.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", lockKey); err != nil {
		slog.Warn("migration lock not released; closing its connection", "error", err)
		_ = conn.Conn().Close(ctx)
	}
}

func appliedVersions(ctx context.Context, conn *pgxpool.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, selectApplied)
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	defer rows.Close()
	applied := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("read schema_migrations: %w", err)
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func apply(ctx context.Context, conn *pgxpool.Conn, m Migration) error {
	stmts := splitStatements(m.SQL)
	if len(stmts) == 0 {
		return errors.New("the file holds no SQL")
	}
	if needsNoTransaction(stmts) {
		for _, s := range stmts {
			if _, err := conn.Exec(ctx, s); err != nil {
				return err
			}
		}
		_, err := conn.Exec(ctx, insertApplied, m.Version)
		return err
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // a no-op once committed
	// No arguments, so pgx sends the whole file as one simple query and the
	// file may hold any number of statements.
	if _, err := tx.Exec(ctx, m.SQL); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, insertApplied, m.Version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// needsNoTransaction reports whether a statement is one Postgres refuses
// inside a transaction block. Only CONCURRENTLY is recognised.
func needsNoTransaction(stmts []string) bool {
	for _, s := range stmts {
		if strings.Contains(strings.ToUpper(s), "CONCURRENTLY") {
			return true
		}
	}
	return false
}

// splitStatements cuts a SQL file into its statements at top-level semicolons,
// leaving a semicolon inside a string, a quoted name or a dollar-quoted body
// alone. Comments are dropped and a statement with nothing but comments is
// omitted, so the result is also what needsNoTransaction should read.
func splitStatements(sql string) []string {
	var (
		out     []string
		cur     strings.Builder
		content bool
	)
	flush := func() {
		if content {
			out = append(out, strings.TrimSpace(cur.String()))
		}
		cur.Reset()
		content = false
	}
	for i := 0; i < len(sql); {
		n, comment := token(sql[i:])
		chunk := sql[i : i+n]
		i += n
		switch {
		case comment:
			cur.WriteByte(' ') // keeps `a/**/b` from becoming `ab`
		case chunk == ";":
			flush()
		default:
			cur.WriteString(chunk)
			if strings.TrimSpace(chunk) != "" {
				content = true
			}
		}
	}
	flush()
	return out
}

// token returns the length of the lexical unit at the start of s that must
// not be cut: a comment, a quoted string or name, or a dollar-quoted body. Any
// other byte is a unit of one. An unterminated unit runs to the end of s.
func token(s string) (n int, comment bool) {
	switch {
	case strings.HasPrefix(s, "--"):
		if j := strings.IndexByte(s, '\n'); j >= 0 {
			return j, true
		}
		return len(s), true
	case strings.HasPrefix(s, "/*"):
		if j := strings.Index(s[2:], "*/"); j >= 0 {
			return j + 4, true
		}
		return len(s), true
	case s[0] == '\'' || s[0] == '"':
		// A doubled quote inside a string reads as two adjacent strings,
		// which cut the same way.
		if j := strings.IndexByte(s[1:], s[0]); j >= 0 {
			return j + 2, false
		}
		return len(s), false
	case s[0] == '$':
		if tag := dollarTag(s); tag != "" {
			if j := strings.Index(s[len(tag):], tag); j >= 0 {
				return len(tag) + j + len(tag), false
			}
			return len(s), false
		}
	}
	return 1, false
}

// dollarTag returns the opening tag of a dollar-quoted string at the start of
// s ($$ or $name$), or "" when s starts with something else, such as $1.
func dollarTag(s string) string {
	for j := 1; j < len(s); j++ {
		c := s[j]
		switch {
		case c == '$':
			return s[:j+1]
		case c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
		case j > 1 && c >= '0' && c <= '9':
		default:
			return ""
		}
	}
	return ""
}

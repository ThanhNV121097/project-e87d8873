package httpx

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgconn"
)

// connectPhase is the text every pgx error of a connection that could not be
// opened begins with, whatever the cause under it: pgx does not export the
// type in this version.
const connectPhase = "failed to connect to"

// Unavailable reports whether err means the database cannot be reached right
// now, however the stack under it says so: a connection that could not be
// opened (refused at the socket, a host name that no longer resolves, or the
// server answering the connection itself with an error such as 55000 "not
// accepting connections", 42501 or 28000), a connection that was reset or
// closed under the query, a request that never left (pgconn's SafeToRetry), or
// a server that is shutting down or not yet accepting (SQLSTATE class 08,
// 57P01 to 57P03). Anything else is a bug or a slow query and retrying does
// not fix it: a failed statement, a deadline (a handler's bound, or a wait for
// a free pool connection) and an abandoned request are all INTERNAL.
//
// Every handler that talks to the database decides 503 against 500 with this
// one predicate: the reads, the owner's writes and the sign-in alike.
func Unavailable(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	// The connect phase is decided before anything under it. A database that
	// refuses to open a connection is a database the service cannot reach: the
	// refusal is a PgError at connect time (pgx wraps it), and reading its
	// SQLSTATE alone would call every such refusal a bug. Even a deadline that
	// expired while connecting is an unreachable database, not a slow statement.
	if strings.Contains(err.Error(), connectPhase) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// A server's own answer to a statement decides on its SQLSTATE alone: a
	// missing table or a constraint is a PgError too, and is not a lost
	// connection.
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return strings.HasPrefix(pg.Code, "08") || pg.Code == "57P01" || pg.Code == "57P02" || pg.Code == "57P03"
	}
	// A network timeout is a deadline in another form.
	var timeout net.Error
	if errors.As(err, &timeout) && timeout.Timeout() {
		return false
	}
	var (
		op  *net.OpError
		dns *net.DNSError
	)
	switch {
	case errors.As(err, &op), errors.As(err, &dns):
		return true
	case errors.Is(err, syscall.ECONNREFUSED), errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF), errors.Is(err, net.ErrClosed):
		return true
	}
	return pgconn.SafeToRetry(err)
}

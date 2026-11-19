package collections

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

// DB is the part of the connection pool the handlers use. server.DB has the
// same method, so the pool satisfies both.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// readTimeout bounds one read: a deadline is a 500, as the contract says.
const readTimeout = 5 * time.Second

const (
	sentenceNoCollection = "That collection does not exist."
	sentenceNoRecord     = "That record does not exist."
	sentenceMethod       = "That method is not offered here."
	sentenceBadQuery     = "The query could not be read."
	sentenceNotReady     = "The service is not ready."
	sentenceInternal     = "Something went wrong. Please try again."
)

// listAnswer is the body of a list. Items is the database's own JSON, passed
// through untouched: a number is never decoded into a float on the way.
type listAnswer struct {
	Items json.RawMessage `json:"items"`
	Total int64           `json:"total"`
	Page  int64           `json:"page"`
	Limit int64           `json:"limit"`
}

// Collection answers GET /v1/{collection}: one page of the collection,
// narrowed, searched and ordered. The collection is checked before the method.
func Collection(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := Lookup(r.PathValue("collection"))
		if !ok {
			httpx.WriteCoded(w, http.StatusNotFound, httpx.CodeNotFound, sentenceNoCollection, "")
			return
		}
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			httpx.WriteCoded(w, http.StatusBadRequest, httpx.CodeInvalidQuery, sentenceBadQuery, "")
			return
		}
		q, qerr := ParseQuery(s, values)
		if qerr != nil {
			httpx.WriteCoded(w, http.StatusBadRequest, httpx.CodeInvalidQuery, qerr.Sentence, qerr.Field)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), readTimeout)
		defer cancel()
		sql, args := BuildList(s, q)
		var (
			total int64
			items []byte
		)
		if err := db.QueryRow(ctx, sql, args...).Scan(&total, &items); err != nil {
			storeFailed(w, r, s.Name, "list", err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, listAnswer{Items: items, Total: total, Page: q.Page, Limit: q.Limit})
	}
}

// Record answers GET /v1/{collection}/{id}: the record, and for a versioned
// collection its version as the ETag. The id matches exactly, capitals
// included.
func Record(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := Lookup(r.PathValue("collection"))
		if !ok {
			httpx.WriteCoded(w, http.StatusNotFound, httpx.CodeNotFound, sentenceNoCollection, "")
			return
		}
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		id := r.PathValue("id")
		if !cleanText(id) {
			// No record can have an id the database cannot store.
			httpx.WriteCoded(w, http.StatusNotFound, httpx.CodeNotFound, sentenceNoRecord, "")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), readTimeout)
		defer cancel()
		var (
			rec     []byte
			version int64
		)
		dest := []any{&rec}
		if s.Versioned {
			dest = append(dest, &version)
		}
		err := db.QueryRow(ctx, BuildRecord(s), id).Scan(dest...)
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteCoded(w, http.StatusNotFound, httpx.CodeNotFound, sentenceNoRecord, "")
			return
		}
		if err != nil {
			storeFailed(w, r, s.Name, "record", err)
			return
		}
		if s.Versioned {
			w.Header().Set("ETag", `"`+strconv.FormatInt(version, 10)+`"`)
		}
		httpx.WriteJSON(w, http.StatusOK, json.RawMessage(rec))
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", http.MethodGet)
	httpx.WriteCoded(w, http.StatusMethodNotAllowed, httpx.CodeMethodNotAllowed, sentenceMethod, "")
}

// storeFailed answers a failed read: 503 only when the database cannot be
// reached (refused, not found, lost) or is going away, 500 for everything else. The real error is
// logged with the request id and the collection, never with the query string:
// a visitor's search words are private. The answer carries one generic
// sentence and nothing of the statement, the host or a constraint.
func storeFailed(w http.ResponseWriter, r *http.Request, collection, op string, err error) {
	attrs := []any{"request_id", httpx.RequestIDFrom(r.Context()), "collection", collection, "op", op, "error", err}
	if errors.Is(err, context.Canceled) {
		slog.Info("read abandoned by the caller", attrs...)
	} else {
		slog.Error("read failed", attrs...)
	}
	if httpx.Unavailable(err) {
		httpx.WriteCoded(w, http.StatusServiceUnavailable, httpx.CodeUnavailable, sentenceNotReady, "")
		return
	}
	httpx.WriteCoded(w, http.StatusInternalServerError, httpx.CodeInternal, sentenceInternal, "")
}

// connectPhase is the text every pgx error of a connection that could not be
// opened begins with, whatever the cause under it: pgx does not export the
// type in this version.
const connectPhase = "failed to connect to"

// unavailable reports whether err means the database cannot be reached right
// now, however the stack under it says so: a connection that could not be
// opened (refused at the socket, a host name that no longer resolves, or the
// server answering the connection itself with an error such as 55000 "not
// accepting connections", 42501 or 28000), a connection that was reset or
// closed under the query, a request that never left (pgconn's SafeToRetry), or
// a server that is shutting down or not yet accepting (SQLSTATE class 08,
// 57P01 to 57P03). Anything else is a bug or a slow query and retrying does
// not fix it: a failed statement, a deadline (the 5 s read bound, or a wait for
// a free pool connection) and an abandoned request are all INTERNAL.
func unavailable(err error) bool {
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

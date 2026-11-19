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
// narrowed, searched and ordered, and POST /v1/{collection}: the owner adds a
// record (guard lets only the signed-in owner through). The collection is
// checked before the method, and the method before the session.
func Collection(db DB, guard Guard) http.HandlerFunc {
	return collection(db, pgWriter{db}, guard)
}

func collection(db DB, wr Writer, guard Guard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := Lookup(r.PathValue("collection"))
		if !ok {
			httpx.WriteCoded(w, http.StatusNotFound, httpx.CodeNotFound, sentenceNoCollection, "")
			return
		}
		if r.Method != http.MethodGet {
			if r.Method == http.MethodPost && s.Writable() {
				guard(addHandler(s, wr)).ServeHTTP(w, r)
				return
			}
			methodNotAllowed(w, s, false)
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
// included. PUT and DELETE are the owner's change and delete of a bike or a
// riding type, behind guard.
func Record(db DB, guard Guard) http.HandlerFunc {
	return record(db, pgWriter{db}, guard)
}

func record(db DB, wr Writer, guard Guard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := Lookup(r.PathValue("collection"))
		if !ok {
			httpx.WriteCoded(w, http.StatusNotFound, httpx.CodeNotFound, sentenceNoCollection, "")
			return
		}
		if r.Method != http.MethodGet {
			if s.Writable() && r.Method == http.MethodPut {
				guard(changeHandler(s, wr)).ServeHTTP(w, r)
				return
			}
			if s.Writable() && r.Method == http.MethodDelete {
				guard(removeHandler(s, wr)).ServeHTTP(w, r)
				return
			}
			methodNotAllowed(w, s, true)
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

// methodNotAllowed answers a method the route does not offer, with the
// methods it does: every collection offers GET, a writable one also POST on
// its list and PUT and DELETE on a record.
func methodNotAllowed(w http.ResponseWriter, s Spec, record bool) {
	allow := http.MethodGet
	switch {
	case s.Writable() && record:
		allow = "GET, PUT, DELETE"
	case s.Writable():
		allow = "GET, POST"
	}
	w.Header().Set("Allow", allow)
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
	if unavailable(err) {
		httpx.WriteCoded(w, http.StatusServiceUnavailable, httpx.CodeUnavailable, sentenceNotReady, "")
		return
	}
	httpx.WriteCoded(w, http.StatusInternalServerError, httpx.CodeInternal, sentenceInternal, "")
}

// unavailable is the one predicate every handler uses to tell a database that
// cannot be reached (503) from a bug or a slow query (500). It lives in httpx,
// beside the sign-in and the owner's writes, which need it too; this name is
// what the read tests pin.
func unavailable(err error) bool { return httpx.Unavailable(err) }

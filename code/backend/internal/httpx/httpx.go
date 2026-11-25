// Package httpx holds what every HTTP handler of the backend shares: the JSON
// answer, the error envelope, the request id and the middleware around the
// router.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// The error codes of the service contract (services.md section 2.3). Consumers
// branch on the code and the status, never on the sentence. A story adds the
// constants it needs.
const (
	CodeNotFound         = "NOT_FOUND"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	CodeInvalidQuery     = "INVALID_QUERY"
	CodeInternal         = "INTERNAL"
	CodeUnavailable      = "UNAVAILABLE"

	CodeUnauthenticated      = "UNAUTHENTICATED"
	CodeRateLimited          = "RATE_LIMITED"
	CodeIDTaken              = "ID_TAKEN"
	CodeStaleRecord          = "STALE_RECORD"
	CodeInUse                = "IN_USE"
	CodePreconditionRequired = "PRECONDITION_REQUIRED"
	CodeMalformedRequest     = "MALFORMED_REQUEST"
	CodeValidationFailed     = "VALIDATION_FAILED"
	CodePayloadTooLarge      = "PAYLOAD_TOO_LARGE"
	CodeBikeUnavailable      = "BIKE_UNAVAILABLE"
)

// RequestIDHeader carries the id of a request, on the way in and on every answer.
const RequestIDHeader = "X-Request-Id"

// WriteJSON answers with v as JSON and the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write the response body", "error", err)
	}
}

type errorEnvelope struct {
	Error string `json:"error"`
	Code  string `json:"code"`
	Field string `json:"field,omitempty"`
}

// WriteCoded answers with the error envelope every non-2xx answer uses:
// {"error": "<one sentence a person can read>", "code": "UPPER_SNAKE"} and, when
// one request field is at fault, "field". The sentence is for the customer or
// the owner, so it never carries an internal message; log that.
func WriteCoded(w http.ResponseWriter, status int, code, sentence, field string) {
	WriteJSON(w, status, errorEnvelope{Error: sentence, Code: code, Field: field})
}

// NoStore forbids every cache between the server and the browser from keeping
// an answer: a saved edit must show at once, on every page (ADMIN-007).
func NoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

type requestIDKey struct{}

// requestIDPattern is what a caller's X-Request-Id must look like to be kept.
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID gives every request an id: the caller's X-Request-Id when it
// matches ^[A-Za-z0-9._-]{1,64}$, otherwise 32 random hex characters. The id
// is set on the response before the handler runs, so every answer carries it,
// a recovered panic included, and it is kept in the request's context for the
// log lines. A header that does not match is replaced, never echoed.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

// RequestIDFrom returns the id RequestID gave the request, or "" outside it.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func newRequestID() string {
	b := make([]byte, 16)
	// crypto/rand.Read cannot fail since Go 1.24: it crashes the program
	// instead of returning an error.
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Recover turns a panic in a handler into a logged 500 with the generic
// sentence, instead of a dropped connection.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("handler panicked", "request_id", RequestIDFrom(r.Context()), "method", r.Method, "path", r.URL.Path, "panic", rec)
				WriteCoded(w, http.StatusInternalServerError, CodeInternal, "Something went wrong. Please try again.", "")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// statusRecorder remembers the status a handler wrote.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Log writes one line per request: request id, method, path (never the query,
// which can carry a visitor's search words), status and duration.
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request", "request_id", RequestIDFrom(r.Context()), "method", r.Method, "path", r.URL.Path, "status", rec.status, "ms", time.Since(start).Milliseconds())
	})
}

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
// not fix it: a failed statement, a deadline (a handler's time bound, or a wait
// for a free pool connection) and an abandoned request are all INTERNAL. Every
// package that answers a store failure calls this one predicate.
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

// Package httpx holds what every HTTP handler of the backend shares: the JSON
// answer, the error envelope, the request id and the middleware around the
// router.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"time"
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

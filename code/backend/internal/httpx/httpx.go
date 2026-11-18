// Package httpx holds what every HTTP handler of the backend shares: the JSON
// answer, the error envelope and the middleware around the router.
package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

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
}

// WriteError answers with the error envelope every non-2xx answer uses:
// {"error": "<one sentence a person can read>"}. The sentence is for the
// customer or the owner, so it never carries an internal message; log that.
func WriteError(w http.ResponseWriter, status int, sentence string) {
	WriteJSON(w, status, errorEnvelope{Error: sentence})
}

// NoStore forbids every cache between the server and the browser from keeping
// an answer: a saved edit must show at once, on every page (ADMIN-007).
func NoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// Recover turns a panic in a handler into a logged 500 with the generic
// sentence, instead of a dropped connection.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("handler panicked", "method", r.Method, "path", r.URL.Path, "panic", rec)
				WriteError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
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

// Log writes one line per request: method, path (never the query, which can
// carry a visitor's search words), status and duration.
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "ms", time.Since(start).Milliseconds())
	})
}

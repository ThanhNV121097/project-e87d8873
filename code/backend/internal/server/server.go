// Package server builds the HTTP handler: every route the backend serves.
//
// Routes carry no /api prefix. In a deployment the edge proxy strips it
// (handle_path /api/* in deploy.yml), so the browser's /api/v1/products is
// this server's /v1/products.
package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

// DB is the part of the connection pool the handlers use; *pgxpool.Pool
// satisfies it. Handlers take this, not the pool, so a test can stand in.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Router returns the whole application as one handler. A story adds its
// routes here with one line each, and takes the DB it needs from db.
func Router(db DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health(db))
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteError(w, http.StatusNotFound, "That address does not exist.")
	})
	return httpx.Recover(httpx.Log(httpx.NoStore(mux)))
}

// health answers 200 only while a query reaches the database. The server
// listens only after its migrations succeeded, so a 200 also means the schema
// is in place; a check that ignored the database would call a broken app
// healthy.
func health(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		var one int
		if err := db.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil || one != 1 {
			slog.Error("health check: the database did not answer", "error", err)
			httpx.WriteError(w, http.StatusServiceUnavailable, "The service is not ready.")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

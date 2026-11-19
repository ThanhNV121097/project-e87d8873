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

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/collections"
	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
	"github.com/ThanhNV121097/project-e87d8873/backend/internal/orders"
)

// DB is the part of the connection pool the handlers use; *pgxpool.Pool
// satisfies it. Handlers take this, not the pool, so a test can stand in.
// Begin is for the one thing that needs a transaction: placing an order.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Router returns the whole application as one handler. A story adds its
// routes here with one line each, and takes the DB it needs from db.
func Router(db DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health(db))
	// No method in these patterns: the handler, not the mux, decides what a
	// method other than GET gets (405 with Allow).
	// The one literal address of the orders story: a longer path under it is
	// the collections' 404, and any method but POST here is a 405.
	mux.HandleFunc("/v1/orders", orders.Place(orders.NewStore(db), time.Now))
	mux.HandleFunc("/v1/{collection}", collections.Collection(db))
	mux.HandleFunc("/v1/{collection}/{id}", collections.Record(db))
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteCoded(w, http.StatusNotFound, httpx.CodeNotFound, "That address does not exist.", "")
	})
	return httpx.RequestID(httpx.Recover(httpx.Log(httpx.NoStore(mux))))
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
			slog.Error("health check: the database did not answer", "request_id", httpx.RequestIDFrom(r.Context()), "error", err)
			httpx.WriteCoded(w, http.StatusServiceUnavailable, httpx.CodeUnavailable, "The service is not ready.", "")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

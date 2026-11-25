// Command api is the backend's only binary: it connects to the database,
// applies every pending migration, and only then listens.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/auth"
	"github.com/ThanhNV121097/project-e87d8873/backend/internal/config"
	"github.com/ThanhNV121097/project-e87d8873/backend/internal/db"
	"github.com/ThanhNV121097/project-e87d8873/backend/internal/migrate"
	"github.com/ThanhNV121097/project-e87d8873/backend/internal/server"
	"github.com/ThanhNV121097/project-e87d8873/backend/migrations"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("backend stopped", "error", err)
		os.Exit(1)
	}
}

// run starts the server in the order that keeps a half-started one from ever
// answering: read the environment, connect, migrate, listen. It returns when
// the server fails or a signal asks it to stop.
func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to the database: %w", err)
	}
	defer pool.Close()

	if err := migrate.Up(ctx, pool, migrations.FS); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	// The owner's two secrets are used here and nowhere else: the account keeps
	// a salted hash and a fingerprint, never the password.
	account := auth.NewAccount(cfg.AdminEmail, cfg.AdminPassword)
	if !account.Configured() {
		slog.Warn("admin account is not configured: nobody can sign in until ADMIN_EMAIL and ADMIN_PASSWORD are set")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Router(pool, account),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	failed := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		failed <- srv.ListenAndServe()
	}()

	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdown)
}

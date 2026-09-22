package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Matyjash/PollAndGo/backend/internal/httpapi"
	"github.com/Matyjash/PollAndGo/backend/internal/poll"
	"github.com/Matyjash/PollAndGo/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Error("database configuration failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	if _, err = db.Exec(ctx, migrations.InitSQL); err != nil {
		log.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	port := envOrDefault("PORT", "8080")
	handler := httpapi.WithCORS(httpapi.New(poll.NewStore(db), log), envOrDefault("ALLOWED_ORIGINS", "http://localhost:5173"))
	server := &http.Server{Addr: ":" + port, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Info("server listening", "port", port)
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/vgartg/goauction/internal/api"
    "github.com/vgartg/goauction/internal/auction"
    "github.com/vgartg/goauction/internal/config"
    "github.com/vgartg/goauction/internal/metrics"
    "github.com/vgartg/goauction/internal/repository"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        slog.Error("failed to load config", "error", err)
        os.Exit(1)
    }

    repo, err := repository.NewPostgresRepo(cfg.DatabaseURL)
    if err != nil {
        slog.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer repo.Close()

    if err := repo.RunMigrations(cfg.DatabaseURL); err != nil {
        slog.Error("failed to run migrations", "error", err)
        os.Exit(1)
    }

    wsManager := api.NewWebSocketManager()
    engine := auction.NewEngine(repo, wsManager)

    // Restore active timers after restart
    lots, err := repo.GetActiveLots(context.Background())
    if err != nil {
        slog.Error("failed to load active lots", "error", err)
    } else {
        metrics.ActiveLots.Set(float64(len(lots)))
        for _, lot := range lots {
            go engine.StartTimerForLot(lot)
        }
    }

    r := chi.NewRouter()
    handlers := api.NewHandlers(engine)
    api.SetupRoutes(r, handlers, wsManager)

    if cfg.MetricsEnabled {
        r.Handle("/metrics", promhttp.Handler())
        slog.Info("metrics endpoint enabled at /metrics")
    }

    server := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: r,
    }

    go func() {
        slog.Info("starting server", "port", cfg.Port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server failed", "error", err)
            os.Exit(1)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    slog.Info("shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        slog.Error("server forced shutdown", "error", err)
    }
    slog.Info("server exited")
}
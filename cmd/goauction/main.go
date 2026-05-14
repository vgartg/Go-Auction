package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/vgartg/goauction/internal/api"
    "github.com/vgartg/goauction/internal/config"
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

    r := chi.NewRouter()
    handlers := api.NewHandlers(repo)
    api.SetupRoutes(r, handlers)

    slog.Info("starting server", "port", cfg.Port)
    if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
        slog.Error("server failed", "error", err)
        os.Exit(1)
    }
}
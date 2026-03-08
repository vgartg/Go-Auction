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
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vgartg/goauction/internal/api"
	"github.com/vgartg/goauction/internal/auction"
	"github.com/vgartg/goauction/internal/auth"
	"github.com/vgartg/goauction/internal/config"
	"github.com/vgartg/goauction/internal/httpx"
	"github.com/vgartg/goauction/internal/metrics"
	"github.com/vgartg/goauction/internal/repository"
	"github.com/vgartg/goauction/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	repo, err := repository.NewPostgresRepo(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	if err := repo.RunMigrations(cfg.DatabaseURL); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	wsManager := api.NewWebSocketManager()
	engine := auction.NewEngine(repo, wsManager, auction.Config{
		SnipingWindow:    cfg.SnipingWindow,
		SnipingExtension: cfg.SnipingExtension,
	})
	authSvc := auth.NewService(repo, cfg.JWTSecret)

	lots, err := repo.GetActiveLots(context.Background())
	if err != nil {
		logger.Error("failed to load active lots", "error", err)
	} else {
		metrics.ActiveLots.Set(float64(len(lots)))
		for _, lot := range lots {
			engine.StartTimerForLot(lot)
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(httpx.AccessLog(logger))

	bidLimiter := httpx.NewRateLimiter("bid", cfg.BidRatePerSec, cfg.BidBurst)
	authLimiter := httpx.NewRateLimiter("auth", cfg.AuthRatePerSec, cfg.AuthBurst)

	apiHandlers := api.NewHandlers(engine, authSvc, repo)
	api.SetupRoutes(r, apiHandlers, wsManager, authSvc, bidLimiter, authLimiter)

	webHandlers := web.NewHandlers(engine, authSvc, repo)
	web.SetupRoutes(r, webHandlers, authSvc, bidLimiter, authLimiter)

	if cfg.MetricsEnabled {
		r.Handle("/metrics", promhttp.Handler())
		logger.Info("metrics endpoint enabled at /metrics")
	}
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting server",
			"port", cfg.Port,
			"sniping_window", cfg.SnipingWindow,
			"sniping_extension", cfg.SnipingExtension)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced shutdown", "error", err)
	}
	logger.Info("server exited")
}

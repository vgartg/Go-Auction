package config

import (
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port           string `env:"PORT" envDefault:"8080"`
	DatabaseURL    string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/goauction?sslmode=disable"`
	MetricsEnabled bool   `env:"METRICS_ENABLED" envDefault:"true"`
	JWTSecret      string `env:"JWT_SECRET" envDefault:"dev-insecure-secret-please-change"`

	// Anti-sniping: if a bid arrives within SnipingWindow before close,
	// extend ClosingAt by SnipingExtension. Set Window to 0 to disable.
	SnipingWindow    time.Duration `env:"SNIPING_WINDOW"    envDefault:"30s"`
	SnipingExtension time.Duration `env:"SNIPING_EXTENSION" envDefault:"30s"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

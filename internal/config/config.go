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

	SnipingWindow    time.Duration `env:"SNIPING_WINDOW"    envDefault:"30s"`
	SnipingExtension time.Duration `env:"SNIPING_EXTENSION" envDefault:"30s"`

	BidRatePerSec  float64 `env:"BID_RATE_PER_SEC"  envDefault:"5"`
	BidBurst       float64 `env:"BID_BURST"         envDefault:"10"`
	AuthRatePerSec float64 `env:"AUTH_RATE_PER_SEC" envDefault:"1"`
	AuthBurst      float64 `env:"AUTH_BURST"        envDefault:"5"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

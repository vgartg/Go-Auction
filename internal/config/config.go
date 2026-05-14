package config

import (
	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port           string `env:"PORT" envDefault:"8080"`
	DatabaseURL    string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/goauction?sslmode=disable"`
	MetricsEnabled bool   `env:"METRICS_ENABLED" envDefault:"true"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

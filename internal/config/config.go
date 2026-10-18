package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	Database      string `env:"POSTGRES_DATABASE"`
	Password      string `env:"POSTGRES_PASSWORD"`
	Host          string `env:"POSTGRES_HOST"`
	Port          string `env:"POSTGRES_PORT"`
	Username      string `env:"POSTGRES_USER"`
	MigrationsDir string `env:"POSTGRES_MIGRATIONS_DIR"`
	SSLMode       string `env:"POSTGRES_SSLMODE"`
}

type Config struct {
	PostgresConfig PostgresConfig
}

func NewConfig() (*Config, error) {
	if err := godotenv.Load("internal/config/.env"); err != nil {
		return nil, err
	}
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

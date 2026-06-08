package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBDriver   string
	DBDSN      string
	JWTSecret  string
	AdminEmail string
	ServerPort string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DBDriver:   getEnv("DB_DRIVER", "sqlite"),
		DBDSN:      getEnv("DB_DSN", "./data/med_book.db?_journal_mode=WAL&_foreign_keys=1"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production"),
		AdminEmail: os.Getenv("ADMIN_EMAIL"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}

	if cfg.DBDriver == "sqlite" && cfg.DBDSN == "" {
		return nil, fmt.Errorf("DB_DSN is required for sqlite")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

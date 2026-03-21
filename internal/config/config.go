package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	BaseURL     string
}

func Load() *Config {
	return &Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

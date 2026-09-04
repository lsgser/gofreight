package config

import (
	"os"
	"strconv"
)

// Environment represents the application runtime environment.
type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Production  Environment = "production"
)

// Config holds application-wide configuration, loaded from environment variables.
type Config struct {
	Environment Environment
	Host        string
	Port        int
	DatabaseURL string
	SecretKey   string
	LogLevel    string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "3000"))

	return &Config{
		Environment: Environment(getEnv("GOFREIGHT_ENV", "development")),
		Host:        getEnv("HOST", "0.0.0.0"),
		Port:        port,
		DatabaseURL: getEnv("DATABASE_URL", "sqlite://db/development.db"),
		SecretKey:   getEnv("SECRET_KEY", "change-me-in-production"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func (c *Config) IsDevelopment() bool { return c.Environment == Development }
func (c *Config) IsProduction() bool  { return c.Environment == Production }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

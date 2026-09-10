package config

/*
|--------------------------------------------------------------------------
| Config
|--------------------------------------------------------------------------
|
| Implements Config as part of the config package in the Gofreight
| framework. Key symbols: Environment, Config, Load, Timezone, Currency,
| Locale.
| 
| The config package loads .env files, resolves MAIL_DRIVER and database
| URLs, and reads config/app.yaml.
| 
| Database drivers and app keys are validated early so misconfiguration
| fails fast at boot.
| 
| Application code reads config through helpers rather than os.Getenv
| scattered across the codebase.
| 
| Symbols defined here include: Environment (exported type); Development
| (exported value); Test (exported value); Production (exported value);
| Config (exported type); Load (Load reads configuration from config files
| and environment variables.); Timezone (Timezone returns the app timezone
| (override via DEFAULT_TIMEZONE).); Currency (Currency returns the
| default ISO 4217 currency code (override via DEFAULT_CURRENCY).).
| 
*/

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
	AppName     string
	AppURL      string
	AppDebug    bool
	Host        string
	Port        int
	Database    DatabaseConfig
	DatabaseURL string // resolved URL for database.Connect
	AppKey      string // application encryption/signing key (APP_KEY)
	SecretKey   string // deprecated alias of AppKey (SECRET_KEY env)
	LogLevel    string
	LogChannel  string
}

// Load reads configuration from config files and environment variables.
func Load() *Config {
	env := Environment(getEnv("GOFREIGHT_ENV", getEnv("APP_ENV", "development")))
	if env == "local" {
		env = Development
	}
	files := LoadFiles(env)
	files.ApplyEnv(map[string]string{
		"PORT":       "port",
		"HOST":       "host",
		"APP_KEY":    "app_key",
		"SECRET_KEY": "secret_key",
		"LOG_LEVEL":  "log_level",
		"APP_NAME":   "app_name",
		"APP_URL":    "app_url",
	})

	port, _ := strconv.Atoi(getEnv("PORT", "5000"))
	if files.GetInt("port", 0) > 0 && getEnv("PORT", "") == "" {
		port = files.GetInt("port", port)
	}
	if port == 0 {
		port = DefaultPort
	}

	db := LoadDatabaseConfig(files)
	appURL := files.GetString("app_url", getEnv("APP_URL", ""))
	if appURL == "" {
		appURL = DefaultAppURL()
	}

	appKey := ResolveAppKey(files)

	return &Config{
		Environment: env,
		AppName:     files.GetString("app_name", getEnv("APP_NAME", "Gofreight")),
		AppURL:      appURL,
		AppDebug:    getEnv("APP_DEBUG", "true") == "true",
		Host:        files.GetString("host", getEnv("HOST", "0.0.0.0")),
		Port:        port,
		Database:    db,
		DatabaseURL: ResolveDatabaseURL(files),
		AppKey:      appKey,
		SecretKey:   appKey,
		LogLevel:    files.GetString("log_level", getEnv("LOG_LEVEL", "info")),
		LogChannel:  getEnv("LOG_CHANNEL", "stack"),
	}
}

// Timezone returns the app timezone (override via DEFAULT_TIMEZONE).
func (c *Config) Timezone() string {
	return getEnv("DEFAULT_TIMEZONE", "UTC")
}

// Currency returns the default ISO 4217 currency code (override via DEFAULT_CURRENCY).
func (c *Config) Currency() string {
	return getEnv("DEFAULT_CURRENCY", "USD")
}

// Locale returns the default locale tag (override via DEFAULT_LOCALE).
func (c *Config) Locale() string {
	return getEnv("DEFAULT_LOCALE", "en")
}

func (c *Config) IsDevelopment() bool { return c.Environment == Development }
func (c *Config) IsProduction() bool  { return c.Environment == Production }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

package config

import (
	"fmt"
	"os"
	"strings"
)

const DefaultPort = 5000

// DefaultAppURL returns the default application URL for local development.
func DefaultAppURL() string {
	return fmt.Sprintf("http://localhost:%d", DefaultPort)
}

// ResolveRedisURL builds a Redis URL from REDIS_URL or REDIS_HOST/PORT/PASSWORD.
func ResolveRedisURL() string {
	if u := cleanEnv(os.Getenv("REDIS_URL")); u != "" {
		return u
	}
	if u := cleanEnv(os.Getenv("UPSTASH_REDIS_URL")); u != "" {
		return u
	}

	host := firstNonEmpty(cleanEnv(os.Getenv("REDIS_HOST")), "127.0.0.1")
	port := firstNonEmpty(cleanEnv(os.Getenv("REDIS_PORT")), "6379")
	pass := cleanEnv(os.Getenv("REDIS_PASSWORD"))

	if pass != "" {
		return fmt.Sprintf("redis://:%s@%s:%s", pass, host, port)
	}
	return fmt.Sprintf("redis://%s:%s", host, port)
}

// ResolveSessionDriver normalizes SESSION_DRIVER (file/database fall back to memory).
func ResolveSessionDriver() string {
	switch strings.ToLower(cleanEnv(os.Getenv("SESSION_DRIVER"))) {
	case "redis":
		return "redis"
	case "file", "database", "":
		return "memory"
	default:
		return strings.ToLower(cleanEnv(os.Getenv("SESSION_DRIVER")))
	}
}

// ResolveQueueConnection normalizes QUEUE_CONNECTION / QUEUE_DRIVER.
func ResolveQueueConnection() string {
	raw := strings.ToLower(firstNonEmpty(
		cleanEnv(os.Getenv("QUEUE_CONNECTION")),
		cleanEnv(os.Getenv("QUEUE_DRIVER")),
	))
	switch raw {
	case "redis":
		return "redis"
	case "sync", "memory", "":
		return "memory"
	default:
		return "memory"
	}
}

// ResolveCacheStore normalizes CACHE_STORE / CACHE_DRIVER.
func ResolveCacheStore() string {
	raw := strings.ToLower(firstNonEmpty(
		cleanEnv(os.Getenv("CACHE_STORE")),
		cleanEnv(os.Getenv("CACHE_DRIVER")),
	))
	switch raw {
	case "redis":
		return "redis"
	case "file", "database", "memory", "":
		return "memory"
	default:
		return raw
	}
}

// ResolveMailMailer returns the active mail transport (log, smtp, sendgrid, …).
func ResolveMailMailer() string {
	return firstNonEmpty(
		cleanEnv(os.Getenv("MAIL_MAILER")),
		cleanEnv(os.Getenv("MAIL_DRIVER")),
		"log",
	)
}

// ResolveFilesystemDisk returns local or s3 (FILESYSTEM_DISK / STORAGE_PROVIDER).
func ResolveFilesystemDisk() string {
	return firstNonEmpty(
		strings.ToLower(cleanEnv(os.Getenv("FILESYSTEM_DISK"))),
		strings.ToLower(cleanEnv(os.Getenv("STORAGE_PROVIDER"))),
		"local",
	)
}

func cleanEnv(v string) string {
	v = strings.TrimSpace(v)
	if v == "null" || v == `""` || v == "''" {
		return ""
	}
	return strings.Trim(v, `"`)
}

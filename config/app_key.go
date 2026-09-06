package config

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

const defaultAppKey = "change-me-in-production"

// GenerateAppKey returns a new application key (base64-encoded 32 bytes).
func GenerateAppKey() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "base64:" + base64.StdEncoding.EncodeToString(raw), nil
}

// ResolveAppKey reads APP_KEY (or legacy SECRET_KEY) from env and config files.
func ResolveAppKey(files *FileConfig) string {
	raw := firstNonEmpty(
		osGet("APP_KEY"),
		osGet("SECRET_KEY"),
		filesGetString(files, "app_key", ""),
		filesGetString(files, "secret_key", ""),
	)
	if raw == "" {
		return defaultAppKey
	}
	return appKeyMaterial(raw)
}

// appKeyMaterial returns key bytes as a string for signing and encryption.
// Keys are stored as base64:... in .env; legacy hex/plain strings still work.
func appKeyMaterial(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "base64:") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(raw, "base64:"))
		if err == nil && len(decoded) > 0 {
			return string(decoded)
		}
	}
	return raw
}

// IsDefaultAppKey reports whether the resolved key is still the scaffold placeholder.
func IsDefaultAppKey(key string) bool {
	return strings.TrimSpace(key) == "" || key == defaultAppKey
}

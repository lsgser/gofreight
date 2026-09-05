package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileConfig holds values loaded from config/*.yaml per environment.
type FileConfig struct {
	Values map[string]any
}

// LoadFiles loads config/{env}.yaml and config/app.yaml (env overrides app).
func LoadFiles(env Environment) *FileConfig {
	cfg := &FileConfig{Values: make(map[string]any)}
	base := "config"
	mergeFile(cfg, filepath.Join(base, "app.yaml"))
	mergeFile(cfg, filepath.Join(base, string(env)+".yaml"))
	return cfg
}

func mergeFile(cfg *FileConfig, path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var data map[string]any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return
	}
	for k, v := range data {
		cfg.Values[k] = v
	}
}

// GetString returns a string config value (file overrides env when set via Apply).
func (c *FileConfig) GetString(key, fallback string) string {
	if v, ok := c.Values[key].(string); ok && v != "" {
		return v
	}
	return getEnv(strings.ToUpper(key), fallback)
}

// GetInt returns an int config value.
func (c *FileConfig) GetInt(key string, fallback int) int {
	if v, ok := c.Values[key].(int); ok {
		return v
	}
	if f, ok := c.Values[key].(float64); ok {
		return int(f)
	}
	return fallback
}

// ApplyEnv sets environment variables from file config when not already set.
func (c *FileConfig) ApplyEnv(mapping map[string]string) {
	for envKey, fileKey := range mapping {
		if os.Getenv(envKey) != "" {
			continue
		}
		if v, ok := c.Values[fileKey].(string); ok && v != "" {
			os.Setenv(envKey, v)
		}
	}
}

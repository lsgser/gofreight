package config_test

import (
	"strings"
	"testing"

	"github.com/lsgser/gofreight/config"
)

func TestGenerateAppKeyFormat(t *testing.T) {
	key, err := config.GenerateAppKey()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, "base64:") {
		t.Fatalf("expected base64: prefix, got %q", key)
	}
	if len(key) < 20 {
		t.Fatalf("key too short: %q", key)
	}
}

func TestResolveAppKeyPrefersAPPKey(t *testing.T) {
	t.Setenv("APP_KEY", "base64:YXBwLWtleS10ZXN0")
	t.Setenv("SECRET_KEY", "legacy-secret")
	got := config.ResolveAppKey(nil)
	if got != "app-key-test" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveAppKeyLegacySecretKey(t *testing.T) {
	t.Setenv("APP_KEY", "")
	t.Setenv("SECRET_KEY", "legacy-hex-secret")
	got := config.ResolveAppKey(nil)
	if got != "legacy-hex-secret" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveAppKeyDefault(t *testing.T) {
	t.Setenv("APP_KEY", "")
	t.Setenv("SECRET_KEY", "")
	got := config.ResolveAppKey(nil)
	if got != "change-me-in-production" {
		t.Fatalf("got %q", got)
	}
}

func TestIsDefaultAppKey(t *testing.T) {
	if !config.IsDefaultAppKey("change-me-in-production") {
		t.Fatal("expected default placeholder")
	}
	if config.IsDefaultAppKey("real-key-material") {
		t.Fatal("expected non-default")
	}
}

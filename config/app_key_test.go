package config_test

/*
|--------------------------------------------------------------------------
| App Key
|--------------------------------------------------------------------------
|
| Test suite for App Key in the config package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
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
| Run with go test ./config/... or go test for this package from the
| framework root.
| 
*/

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

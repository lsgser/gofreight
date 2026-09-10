package integrations_test

/*
|--------------------------------------------------------------------------
| Wire
|--------------------------------------------------------------------------
|
| Test suite for Wire in the integrations package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| Integrations register pluggable drivers for mail, storage, cache, queue,
| and custom third-party APIs.
| 
| Active() resolves the configured implementation from environment
| variables; wire Application in ConfigureIntegrations.
| 
| Built-in connectors cover SMTP, SendGrid, S3-compatible storage, and
| Redis without vendor-specific SDKs in app code.
| 
| Run with go test ./integrations/... or go test for this package from the
| framework root.
| 
*/

import (
	"testing"

	"github.com/lsgser/gofreight/integrations"
)

type mockEnv map[string]string

func (m mockEnv) Get(key string) string { return m[key] }

func TestBuildServicesLogMailer(t *testing.T) {
	svc, err := integrations.BuildServices(mockEnv{})
	if err != nil {
		t.Fatal(err)
	}
	if svc.Mail == nil || svc.Cache == nil {
		t.Fatal("expected mail and cache")
	}
}

func TestStorageNotConfigured(t *testing.T) {
	_ = integrations.ConfigureAll(mockEnv{})
	s, ok := integrations.Get("storage")
	if !ok {
		t.Fatal("storage should be registered")
	}
	if s.Enabled() {
		t.Fatal("storage should not be enabled without bucket")
	}
}

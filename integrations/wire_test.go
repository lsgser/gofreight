package integrations_test

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

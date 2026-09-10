package integrations

/*
|--------------------------------------------------------------------------
| Drivers
|--------------------------------------------------------------------------
|
| Test suite for Drivers in the integrations package.
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
	"os"
	"testing"
)

type mockPayment struct {
	enabled bool
}

func (m *mockPayment) Name() string { return "mockpay" }

func (m *mockPayment) Configure(env EnvReader) error {
	m.enabled = env.Get("MOCKPAY_API_KEY") != ""
	return nil
}

func (m *mockPayment) Enabled() bool { return m.enabled }

func (m *mockPayment) Category() Category { return CategoryPayment }

func (m *mockPayment) DriverID() string { return "mockpay" }

func TestActiveStorageByProvider(t *testing.T) {
	t.Setenv("STORAGE_PROVIDER", "s3")
	t.Setenv("STORAGE_BUCKET", "my-bucket")
	t.Setenv("STORAGE_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "key")

	s, ok := ActiveStorage(OsEnv{})
	if !ok {
		t.Fatal("expected storage driver")
	}
	if s.DriverID() != "s3" {
		t.Fatalf("got driver %s", s.DriverID())
	}
}

func TestActiveStorageFilesystemDisk(t *testing.T) {
	t.Setenv("FILESYSTEM_DISK", "s3")
	t.Setenv("AWS_BUCKET", "my-bucket")
	t.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "key")

	s, ok := ActiveStorage(OsEnv{})
	if !ok {
		t.Fatal("expected storage driver")
	}
	if s.DriverID() != "s3" {
		t.Fatalf("got driver %s", s.DriverID())
	}
}

func TestActiveEmailMailMailer(t *testing.T) {
	t.Setenv("MAIL_MAILER", "sendgrid")
	t.Setenv("MAIL_FROM_ADDRESS", "noreply@example.com")
	t.Setenv("SENDGRID_API_KEY", "sg-key")

	e, ok := ActiveEmail(OsEnv{})
	if !ok {
		t.Fatal("expected email driver")
	}
	if e.DriverID() != "sendgrid" {
		t.Fatalf("got driver %s", e.DriverID())
	}
}

func TestActivePaymentCustomDriver(t *testing.T) {
	Register("mockpay", func() Integration { return &mockPayment{} })
	t.Setenv("PAYMENT_PROVIDER", "mockpay")
	t.Setenv("MOCKPAY_API_KEY", "key")

	i, ok := ActivePayment(OsEnv{})
	if !ok {
		t.Fatal("expected custom payment driver")
	}
	cd, ok := AsCategorizedDriver(i)
	if !ok || cd.DriverID() != "mockpay" {
		t.Fatalf("unexpected driver: %v", i)
	}
}

func TestActivePaymentNoneWithoutRegistration(t *testing.T) {
	os.Unsetenv("PAYMENT_PROVIDER")
	_, ok := ActivePayment(OsEnv{})
	if ok {
		t.Fatal("expected no payment driver without registration")
	}
}

func TestActiveRedisCacheStore(t *testing.T) {
	t.Setenv("CACHE_STORE", "redis")
	t.Setenv("REDIS_HOST", "127.0.0.1")
	t.Setenv("REDIS_PORT", "6379")

	r, ok := ActiveRedis(OsEnv{})
	if !ok {
		t.Fatal("expected redis cache driver")
	}
	if r.DriverID() != "redis" {
		t.Fatalf("got %s", r.DriverID())
	}
}

func TestProviderEnvKeys(t *testing.T) {
	if ProviderEnvKey(CategoryMail) != "EMAIL_PROVIDER" {
		t.Fatal("mail env key mismatch")
	}
	if ProviderEnvKey(CategoryCache) != "CACHE_DRIVER" {
		t.Fatal("cache env key mismatch")
	}
	if ProviderEnvKey(CategoryPayment) != "PAYMENT_PROVIDER" {
		t.Fatal("payment env key mismatch")
	}
}

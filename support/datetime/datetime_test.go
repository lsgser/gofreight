package datetime_test

/*
|--------------------------------------------------------------------------
| Datetime
|--------------------------------------------------------------------------
|
| Test suite for Datetime in the datetime package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| Datetime helpers parse and format timestamps consistently across models,
| APIs, and views.
| 
| Complements carbon-style usage documented in configuration and ORM
| guides.
| 
| Run with go test ./support/datetime/... or go test for this package from
| the framework root.
| 
*/

import (
	"testing"
	"time"

	"github.com/lsgser/gofreight/support/datetime"
)

func TestNowAndFormat(t *testing.T) {
	now := datetime.Now()
	if now.IsZero() {
		t.Fatal("expected non-zero now")
	}
	if datetime.DateTimeString(now) == "" {
		t.Fatal("expected datetime string")
	}
}

func TestFromTimeRoundTrip(t *testing.T) {
	std := time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC)
	c := datetime.FromTime(std, "UTC")
	got := datetime.ToTime(c)
	if !got.Equal(std) {
		t.Fatalf("got %v want %v", got, std)
	}
}

func TestParse(t *testing.T) {
	c := datetime.Parse("2026-01-02 15:04:05")
	if datetime.DateString(c) != "2026-01-02" {
		t.Fatalf("got %q", datetime.DateString(c))
	}
}

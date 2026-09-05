package datetime_test

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

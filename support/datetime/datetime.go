// Package datetime provides fluent date helpers for Gofreight apps via github.com/dromara/carbon/v2.
package datetime

import (
	"time"

	carbon "github.com/dromara/carbon/v2"
)

// Carbon is the underlying Carbon instance (github.com/dromara/carbon/v2).
type Carbon = carbon.Carbon

// Now returns the current time.
func Now(timezone ...string) *Carbon {
	return carbon.Now(timezone...)
}

// Parse parses a datetime string.
func Parse(value string, timezone ...string) *Carbon {
	return carbon.Parse(value, timezone...)
}

// Yesterday returns yesterday's date.
func Yesterday(timezone ...string) *Carbon {
	return carbon.Yesterday(timezone...)
}

// Tomorrow returns tomorrow's date.
func Tomorrow(timezone ...string) *Carbon {
	return carbon.Tomorrow(timezone...)
}

// FromTime wraps a standard library time.Time.
func FromTime(t time.Time, timezone ...string) *Carbon {
	return carbon.CreateFromStdTime(t, timezone...)
}

// ToTime converts Carbon to time.Time.
func ToTime(c *Carbon) time.Time {
	if c == nil {
		return time.Time{}
	}
	return c.StdTime()
}

// DateString returns Y-m-d (for date columns).
func DateString(c *Carbon) string {
	if c == nil {
		return ""
	}
	return c.ToDateString()
}

// DateTimeString returns Y-m-d H:i:s (for datetime columns).
func DateTimeString(c *Carbon) string {
	if c == nil {
		return ""
	}
	return c.ToDateTimeString()
}

// ISO8601 returns an ISO-8601 string (RFC3339-style).
func ISO8601(c *Carbon) string {
	if c == nil {
		return ""
	}
	return c.ToIso8601String()
}

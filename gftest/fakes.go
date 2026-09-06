package gftest

import (
	"sync"
	"testing"

	"github.com/lsgser/gofreight/cache"
	"github.com/lsgser/gofreight/jobs"
	"github.com/lsgser/gofreight/mail"
)

// Fakes holds test doubles for framework services (mail, cache, queue).
type Fakes struct {
	Mail  *mail.LogMailer
	Cache *cache.Store
	Queue *jobs.Queue
}

var (
	defaultFakes *Fakes
	fakesOnce      sync.Once
)

// UseFakes initializes framework fakes for testing.
func UseFakes() *Fakes {
	fakesOnce = sync.Once{} // reset for tests
	f := &Fakes{
		Mail:  mail.NewLogMailer(),
		Cache: cache.New(),
		Queue: jobs.New(),
	}
	defaultFakes = f
	return f
}

// Fakes returns the current test fakes instance.
func GetFakes() *Fakes {
	if defaultFakes == nil {
		return UseFakes()
	}
	return defaultFakes
}

// AssertMailSent asserts an email was sent.
func AssertMailSent(t *testing.T, count int) {
	t.Helper()
	Expect(GetFakes().Mail.Count()).Bind(t).ToEqual(count)
}

// AssertJobDispatched asserts jobs were queued.
func AssertJobDispatched(t *testing.T, count int) {
	t.Helper()
	Expect(GetFakes().Queue.Pending()).Bind(t).ToEqual(count)
}

// AssertCached asserts a cache key exists.
func AssertCached(t *testing.T, key string) {
	t.Helper()
	Expect(GetFakes().Cache.Has(key)).Bind(t).ToBeTrue()
}

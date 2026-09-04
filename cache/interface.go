package cache

import "time"

// Cacher is the cache interface (memory, Redis, etc.).
type Cacher interface {
	Get(key string) (any, bool)
	Put(key string, value any, ttl ...time.Duration)
	Has(key string) bool
	Forget(key string)
	Flush()
	Remember(key string, ttl time.Duration, fn func() any) any
}

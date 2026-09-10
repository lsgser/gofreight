package cache

/*
|--------------------------------------------------------------------------
| Interface
|--------------------------------------------------------------------------
|
| Implements Interface as part of the cache package in the Gofreight
| framework. Key symbols: Cacher.
| 
| The cache package defines CacheStore implementations: in-memory, file,
| Redis, and HTTP cache middleware.
| 
| Application wiring selects the driver from CACHE_STORE and related env
| vars via integrations.
| 
| Use cache for rate limiting data, session alternatives, or fragment
| caching in controllers.
| 
| Symbols defined here include: Cacher (exported type).
| 
*/

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

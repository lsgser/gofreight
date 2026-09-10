package cache

/*
|--------------------------------------------------------------------------
| Cache
|--------------------------------------------------------------------------
|
| Implements Cache as part of the cache package in the Gofreight
| framework. Key symbols: Store, New, Put, Get, Has, Forget.
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
| Symbols defined here include: Store (exported type); New (New creates an
| in-memory cache store.); Put (Put stores a value with optional TTL.);
| Get (Get retrieves a value.); Has (Has returns true if the key exists
| and is not expired.); Forget (Forget removes a key.); Flush (Flush
| clears all keys.); Remember (Remember gets or computes and stores a
| value.).
| 
*/

import (
	"sync"
	"time"
)

// Store is an in-memory cache with optional TTL.
type Store struct {
	mu    sync.RWMutex
	items map[string]entry
}

type entry struct {
	value      any
	expiration time.Time
}

// New creates an in-memory cache store.
func New() *Store {
	return &Store{items: make(map[string]entry)}
}

// Put stores a value with optional TTL.
func (s *Store) Put(key string, value any, ttl ...time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := entry{value: value}
	if len(ttl) > 0 && ttl[0] > 0 {
		e.expiration = time.Now().Add(ttl[0])
	}
	s.items[key] = e
}

// Get retrieves a value.
func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.items[key]
	if !ok {
		return nil, false
	}
	if !e.expiration.IsZero() && time.Now().After(e.expiration) {
		return nil, false
	}
	return e.value, true
}

// Has returns true if the key exists and is not expired.
func (s *Store) Has(key string) bool {
	_, ok := s.Get(key)
	return ok
}

// Forget removes a key.
func (s *Store) Forget(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
}

// Flush clears all keys.
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]entry)
}

// Remember gets or computes and stores a value.
func (s *Store) Remember(key string, ttl time.Duration, fn func() any) any {
	if v, ok := s.Get(key); ok {
		return v
	}
	v := fn()
	s.Put(key, v, ttl)
	return v
}

// Forever stores without expiration.
func (s *Store) Forever(key string, value any) {
	s.Put(key, value)
}

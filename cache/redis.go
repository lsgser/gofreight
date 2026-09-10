package cache

/*
|--------------------------------------------------------------------------
| Redis
|--------------------------------------------------------------------------
|
| Implements Redis as part of the cache package in the Gofreight
| framework. Key symbols: RedisStore, NewRedis, Get, Put, Has, Forget.
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
| Symbols defined here include: RedisStore (exported type); NewRedis
| (NewRedis creates a Redis-backed cache.); Client (Client returns the
| underlying Redis client for pub/sub and other uses.).
| 
*/

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements Cacher using Redis.
type RedisStore struct {
	client *redis.Client
	prefix string
}

// NewRedis creates a Redis-backed cache.
func NewRedis(url, prefix string) (*RedisStore, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	if prefix == "" {
		prefix = "gofreight:"
	}
	return &RedisStore{client: client, prefix: prefix}, nil
}

func (s *RedisStore) key(k string) string { return s.prefix + k }

func (s *RedisStore) Get(key string) (any, bool) {
	val, err := s.client.Get(context.Background(), s.key(key)).Result()
	if err != nil {
		return nil, false
	}
	var out any
	if err := json.Unmarshal([]byte(val), &out); err != nil {
		return val, true
	}
	return out, true
}

func (s *RedisStore) Put(key string, value any, ttl ...time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	exp := time.Duration(0)
	if len(ttl) > 0 {
		exp = ttl[0]
	}
	s.client.Set(context.Background(), s.key(key), data, exp)
}

func (s *RedisStore) Has(key string) bool {
	_, ok := s.Get(key)
	return ok
}

func (s *RedisStore) Forget(key string) {
	s.client.Del(context.Background(), s.key(key))
}

func (s *RedisStore) Flush() {
	s.client.FlushDB(context.Background())
}

func (s *RedisStore) Remember(key string, ttl time.Duration, fn func() any) any {
	if v, ok := s.Get(key); ok {
		return v
	}
	v := fn()
	s.Put(key, v, ttl)
	return v
}

// Client returns the underlying Redis client for pub/sub and other uses.
func (s *RedisStore) Client() *redis.Client { return s.client }

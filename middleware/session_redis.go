package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisSessionStore stores sessions in Redis for multi-instance deployments.
type RedisSessionStore struct {
	client *redis.Client
	prefix string
}

// NewRedisSessionStore creates a Redis-backed session store.
func NewRedisSessionStore(url, prefix string) (*RedisSessionStore, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return &RedisSessionStore{
		client: redis.NewClient(opts),
		prefix: prefix,
	}, nil
}

func (r *RedisSessionStore) key(id string) string {
	return r.prefix + id
}

func (r *RedisSessionStore) Get(id string) (*Session, bool) {
	ctx := context.Background()
	raw, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		return nil, false
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, false
	}
	s := &Session{id: id, data: data}
	return s, true
}

func (r *RedisSessionStore) Save(session *Session, maxAge time.Duration) {
	ctx := context.Background()
	session.mu.RLock()
	data := make(map[string]any, len(session.data))
	for k, v := range session.data {
		data[k] = v
	}
	session.mu.RUnlock()
	raw, _ := json.Marshal(data)
	r.client.Set(ctx, r.key(session.id), raw, maxAge)
}

func (r *RedisSessionStore) Delete(id string) {
	r.client.Del(context.Background(), r.key(id))
}

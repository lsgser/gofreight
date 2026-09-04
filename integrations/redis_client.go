package integrations

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// redis field on Redis struct
func (r *Redis) ensureClient() *redis.Client {
	if r.client != nil {
		return r.client
	}
	if !r.enabled || r.URL == "" {
		return nil
	}
	opts, err := redis.ParseURL(r.URL)
	if err != nil {
		return nil
	}
	r.client = redis.NewClient(opts)
	return r.client
}

// Ping verifies Redis connectivity.
func (r *Redis) Ping(ctx context.Context) error {
	c := r.ensureClient()
	if c == nil {
		return ErrNotConfigured
	}
	return c.Ping(ctx).Err()
}

// GetClient returns the underlying Redis client.
func (r *Redis) GetClient() *redis.Client {
	return r.ensureClient()
}

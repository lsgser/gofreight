package integrations

/*
|--------------------------------------------------------------------------
| Redis Client
|--------------------------------------------------------------------------
|
| Implements Redis Client as part of the integrations package in the
| Gofreight framework. Key symbols: Ping, GetClient.
| 
| Integrations register pluggable drivers for mail, storage, cache, queue,
| and custom third-party APIs.
| 
| Active() resolves the configured implementation from environment
| variables; wire Application in ConfigureIntegrations.
| 
| Built-in connectors cover SMTP, SendGrid, S3-compatible storage, and
| Redis without vendor-specific SDKs in app code.
| 
| Symbols defined here include: Ping (Ping verifies Redis connectivity.);
| GetClient (GetClient returns the underlying Redis client.).
| 
*/

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

package config_test

import (
	"testing"

	"github.com/lsgser/gofreight/config"
)

func TestResolveRedisURLFromParts(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_HOST", "10.0.0.5")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_PASSWORD", "secret")
	got := config.ResolveRedisURL()
	want := "redis://:secret@10.0.0.5:6380"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveQueueConnectionSync(t *testing.T) {
	t.Setenv("QUEUE_CONNECTION", "sync")
	t.Setenv("QUEUE_DRIVER", "")
	if config.ResolveQueueConnection() != "memory" {
		t.Fatal("expected memory")
	}
}

func TestResolveCacheStoreRedis(t *testing.T) {
	t.Setenv("CACHE_STORE", "redis")
	if config.ResolveCacheStore() != "redis" {
		t.Fatal("expected redis")
	}
}

func TestDefaultPortConstant(t *testing.T) {
	if config.DefaultPort != 5000 {
		t.Fatalf("expected 5000 got %d", config.DefaultPort)
	}
}

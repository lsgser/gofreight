package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// HTTPCacheMiddleware adds Cache-Control and ETag support.
func HTTPCacheMiddleware(ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Cache-Control", "public, max-age="+itoa(int(ttl.Seconds())))
			next.ServeHTTP(w, r)
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// FragmentCache caches rendered template fragments.
type FragmentCache struct {
	store Cacher
}

// NewFragmentCache creates a fragment cache wrapper.
func NewFragmentCache(store Cacher) *FragmentCache {
	return &FragmentCache{store: store}
}

// Remember renders or returns a cached HTML fragment.
func (f *FragmentCache) Remember(key string, ttl time.Duration, render func() (string, error)) (string, error) {
	if v, ok := f.store.Get(key); ok {
		if s, ok := v.(string); ok {
			return s, nil
		}
	}
	html, err := render()
	if err != nil {
		return "", err
	}
	f.store.Put(key, html, ttl)
	return html, nil
}

// FragmentKey hashes a cache key from parts.
func FragmentKey(parts ...string) string {
	h := sha256.Sum256([]byte(joinParts(parts)))
	return "fragment:" + hex.EncodeToString(h[:8])
}

func joinParts(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "|"
		}
		out += p
	}
	return out
}

// RememberHTTP caches full HTTP responses by URL path (use with caution).
func RememberHTTP(store Cacher, ttl time.Duration, key string, fn func() ([]byte, error)) ([]byte, error) {
	if v, ok := store.Get(key); ok {
		if b, ok := v.([]byte); ok {
			return b, nil
		}
	}
	body, err := fn()
	if err != nil {
		return nil, err
	}
	store.Put(key, body, ttl)
	return body, nil
}

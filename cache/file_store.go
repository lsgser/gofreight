package cache

/*
|--------------------------------------------------------------------------
| File Store
|--------------------------------------------------------------------------
|
| Implements File Store as part of the cache package in the Gofreight
| framework. Key symbols: FileStore, NewFileStore, Put, Get, Has, Forget.
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
| Symbols defined here include: FileStore (exported type); NewFileStore
| (NewFileStore creates a file-backed cache store.).
| 
*/

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileStore persists cache entries as JSON files on disk.
type FileStore struct {
	dir string
	mu  sync.Mutex
}

type fileEntry struct {
	Value      json.RawMessage `json:"value"`
	Expiration int64           `json:"expiration,omitempty"`
}

// NewFileStore creates a file-backed cache store.
func NewFileStore(dir string) (*FileStore, error) {
	if dir == "" {
		dir = filepath.Join("storage", "framework", "cache", "data")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FileStore{dir: dir}, nil
}

func (f *FileStore) path(key string) string {
	name := sanitizeCacheKey(key)
	return filepath.Join(f.dir, name+".json")
}

func sanitizeCacheKey(key string) string {
	out := make([]byte, 0, len(key))
	for i := 0; i < len(key); i++ {
		c := key[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_':
			out = append(out, c)
		default:
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "key"
	}
	return string(out)
}

func (f *FileStore) Put(key string, value any, ttl ...time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entry := fileEntry{}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	entry.Value = raw
	if len(ttl) > 0 && ttl[0] > 0 {
		entry.Expiration = time.Now().Add(ttl[0]).Unix()
	}
	data, _ := json.Marshal(entry)
	_ = os.WriteFile(f.path(key), data, 0644)
}

func (f *FileStore) Get(key string) (any, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	raw, err := os.ReadFile(f.path(key))
	if err != nil {
		return nil, false
	}
	var entry fileEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, false
	}
	if entry.Expiration > 0 && time.Now().Unix() > entry.Expiration {
		_ = os.Remove(f.path(key))
		return nil, false
	}
	var value any
	if err := json.Unmarshal(entry.Value, &value); err != nil {
		return nil, false
	}
	return value, true
}

func (f *FileStore) Has(key string) bool {
	_, ok := f.Get(key)
	return ok
}

func (f *FileStore) Forget(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_ = os.Remove(f.path(key))
}

func (f *FileStore) Flush() {
	f.mu.Lock()
	defer f.mu.Unlock()
	entries, _ := os.ReadDir(f.dir)
	for _, e := range entries {
		if !e.IsDir() {
			_ = os.Remove(filepath.Join(f.dir, e.Name()))
		}
	}
}

func (f *FileStore) Remember(key string, ttl time.Duration, fn func() any) any {
	if v, ok := f.Get(key); ok {
		return v
	}
	v := fn()
	f.Put(key, v, ttl)
	return v
}

package middleware

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileSessionStore persists sessions as JSON files on disk.
type FileSessionStore struct {
	dir string
	mu  sync.Mutex
}

// NewFileSessionStore creates a file-backed session store.
func NewFileSessionStore(dir string) (*FileSessionStore, error) {
	if dir == "" {
		dir = filepath.Join("storage", "framework", "sessions")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &FileSessionStore{dir: dir}, nil
}

type fileSessionData struct {
	Data map[string]any `json:"data"`
}

func (f *FileSessionStore) path(id string) string {
	return filepath.Join(f.dir, id+".json")
}

func (f *FileSessionStore) Get(id string) (*Session, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	raw, err := os.ReadFile(f.path(id))
	if err != nil {
		return nil, false
	}
	var stored fileSessionData
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, false
	}
	s := &Session{data: stored.Data, id: id}
	if s.data == nil {
		s.data = make(map[string]any)
	}
	return s, true
}

func (f *FileSessionStore) Save(session *Session, maxAge time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if session.data == nil {
		session.data = make(map[string]any)
	}
	raw, err := json.Marshal(fileSessionData{Data: session.data})
	if err != nil {
		return
	}
	_ = os.WriteFile(f.path(session.id), raw, 0644)
}

func (f *FileSessionStore) Delete(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_ = os.Remove(f.path(id))
}

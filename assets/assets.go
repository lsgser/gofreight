package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Pipeline serves and fingerprints static assets from a public directory.
type Pipeline struct {
	Root      string
	Prefix    string
	mu        sync.RWMutex
	digestMap map[string]string
}

// New creates an asset pipeline rooted at the public directory.
func New(root string) *Pipeline {
	return &Pipeline{
		Root:      root,
		Prefix:    "/assets",
		digestMap: make(map[string]string),
	}
}

// Precompile builds a digest map for all files in the public directory.
func (p *Pipeline) Precompile() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return filepath.WalkDir(p.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(p.Root, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(content)
		digest := hex.EncodeToString(hash[:8])
		ext := filepath.Ext(rel)
		base := strings.TrimSuffix(rel, ext)
		p.digestMap[rel] = fmt.Sprintf("%s-%s%s", base, digest, ext)
		return nil
	})
}

// Path returns the asset URL path, with digest in production mode.
func (p *Pipeline) Path(assetPath string) string {
	assetPath = strings.TrimPrefix(assetPath, "/")

	p.mu.RLock()
	digested, ok := p.digestMap[assetPath]
	p.mu.RUnlock()

	if ok {
		return p.Prefix + "/" + digested
	}
	return p.Prefix + "/" + assetPath
}

// Handler returns an http.Handler that serves static files.
func (p *Pipeline) Handler() http.Handler {
	fileServer := http.FileServer(http.Dir(p.Root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, p.Prefix+"/")
		if path == "" {
			http.NotFound(w, r)
			return
		}

		// Strip digest from filename: app.css -> app-<hash>.css
		servePath := p.resolveDigest(path)
		r.URL.Path = "/" + servePath
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		fileServer.ServeHTTP(w, r)
	})
}

func (p *Pipeline) resolveDigest(path string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for original, digested := range p.digestMap {
		if digested == path {
			return original
		}
	}
	return path
}

// Mount registers the asset handler on a router at the configured prefix.
func (p *Pipeline) Mount(mux *http.ServeMux) {
	mux.Handle(p.Prefix+"/", p.Handler())
}

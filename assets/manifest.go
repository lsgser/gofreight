package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Manifest reads a Vite/webpack manifest for hashed asset paths.
type Manifest struct {
	entries map[string]string
}

// LoadManifest loads public/manifest.json (Vite build output).
func LoadManifest(path string) (*Manifest, error) {
	if path == "" {
		path = filepath.Join("public", "manifest.json")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return &Manifest{entries: map[string]string{}}, nil
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	entries := make(map[string]string)
	for k, v := range data {
		switch val := v.(type) {
		case string:
			entries[k] = val
		case map[string]any:
			if file, ok := val["file"].(string); ok {
				entries[k] = file
			}
		}
	}
	return &Manifest{entries: entries}, nil
}

// Path returns the hashed asset path from the manifest.
func (m *Manifest) Path(entry string) string {
	if m == nil {
		return entry
	}
	if p, ok := m.entries[entry]; ok {
		return "/assets/" + p
	}
	return "/assets/" + entry
}

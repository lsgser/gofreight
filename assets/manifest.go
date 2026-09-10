package assets

/*
|--------------------------------------------------------------------------
| Manifest
|--------------------------------------------------------------------------
|
| Implements Manifest as part of the assets package in the Gofreight
| framework. Key symbols: Manifest, LoadManifest, Path.
| 
| The assets package connects your public/ directory and optional Vite dev
| server to HTTP handlers.
| 
| In development, Vite can proxy hot module replacement; in production, a
| manifest maps entry points to built files.
| 
| GFT templates use #vite and /assets/ paths documented in the templating
| guide.
| 
| Symbols defined here include: Manifest (exported type); LoadManifest
| (LoadManifest loads public/manifest.json (Vite build output).); Path
| (Path returns the hashed asset path from the manifest.).
| 
*/

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

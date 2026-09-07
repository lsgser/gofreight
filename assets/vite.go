package assets

import (
	"fmt"
	"os"
	"strings"
)

// Vite configures the Vite dev server and production manifest.
type Vite struct {
	DevServerURL string
	Manifest     *Manifest
	HotFile      string
}

// NewVite creates Vite asset helpers.
func NewVite() *Vite {
	url := strings.TrimSpace(os.Getenv("VITE_DEV_SERVER_URL"))
	if url == "" {
		url = "http://localhost:5173"
	}
	return &Vite{
		DevServerURL: url,
		HotFile:      "public/hot",
	}
}

// IsRunning reports whether the Vite dev server is active (public/hot exists).
func (v *Vite) IsRunning() bool {
	if v == nil {
		return false
	}
	_, err := os.Stat(v.HotFile)
	return err == nil
}

// Asset returns a script or stylesheet URL for an entry point.
func (v *Vite) Asset(entry string) string {
	if v.IsRunning() {
		return fmt.Sprintf("%s/@vite/%s", strings.TrimSuffix(v.DevServerURL, "/"), entry)
	}
	if v.Manifest != nil {
		return v.Manifest.Path(entry)
	}
	return "/assets/" + entry
}

// ClientTags returns HTML script/link tags for a Vite entry in development or production.
func (v *Vite) ClientTags(entry string) string {
	if v.IsRunning() {
		base := strings.TrimSuffix(v.DevServerURL, "/")
		return fmt.Sprintf(`<script type="module" src="%s/@vite/client"></script>`+"\n"+`<script type="module" src="%s/%s"></script>`, base, base, entry)
	}
	path := "/assets/" + entry
	if v.Manifest != nil {
		path = v.Manifest.Path(entry)
	}
	if strings.HasSuffix(entry, ".css") {
		return fmt.Sprintf(`<link rel="stylesheet" href="%s">`, path)
	}
	return fmt.Sprintf(`<script type="module" src="%s"></script>`, path)
}

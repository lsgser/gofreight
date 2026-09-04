package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SaveFile saves an uploaded file to destDir with a sanitized filename.
func SaveFile(file multipart.File, header *multipart.FileHeader, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	name := sanitizeFilename(header.Filename)
	dest := filepath.Join(destDir, name)

	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}
	return dest, nil
}

// ParseMultipart parses multipart form and returns the first file for a field.
func ParseMultipart(r *http.Request, field string, maxMB int64) (multipart.File, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(maxMB << 20); err != nil {
		return nil, nil, err
	}
	file, header, err := r.FormFile(field)
	if err != nil {
		return nil, nil, fmt.Errorf("missing file field %q", field)
	}
	return file, header, nil
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	if name == "" || name == "." {
		return "upload"
	}
	return name
}

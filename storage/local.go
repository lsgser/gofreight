package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// LocalDisk stores files on the local filesystem (Laravel local disk).
type LocalDisk struct {
	Root string
}

// NewLocalDisk creates a local storage disk.
func NewLocalDisk(root string) (*LocalDisk, error) {
	if root == "" {
		root = "storage/app"
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	return &LocalDisk{Root: root}, nil
}

func (d *LocalDisk) fullPath(key string) string {
	key = filepath.Clean(key)
	key = filepath.Join(d.Root, key)
	return key
}

// Put writes content to the disk.
func (d *LocalDisk) Put(ctx context.Context, key string, body io.Reader, size int64) error {
	_ = ctx
	path := d.fullPath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, body)
	return err
}

// Get reads a file from the disk.
func (d *LocalDisk) Get(ctx context.Context, key string) ([]byte, error) {
	_ = ctx
	return os.ReadFile(d.fullPath(key))
}

// Delete removes a file from the disk.
func (d *LocalDisk) Delete(ctx context.Context, key string) error {
	_ = ctx
	return os.Remove(d.fullPath(key))
}

// Exists reports whether a file exists.
func (d *LocalDisk) Exists(ctx context.Context, key string) bool {
	_ = ctx
	_, err := os.Stat(d.fullPath(key))
	return err == nil
}

// Path returns the absolute filesystem path for a key.
func (d *LocalDisk) Path(key string) string {
	return d.fullPath(key)
}

// URL returns a public URL path (app must serve files from public disk separately).
func (d *LocalDisk) URL(key string) string {
	return "/storage/" + filepath.ToSlash(key)
}

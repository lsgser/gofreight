package controller_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/router"
)

func TestDownloadAndFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.txt")
	if err := os.WriteFile(path, []byte("hello file"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("download", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		base := controller.NewBase(rec, req)
		if err := base.Download(path, "export.txt"); err != nil {
			t.Fatal(err)
		}
		if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="export.txt"` {
			t.Fatalf("unexpected disposition %q", got)
		}
		if rec.Body.String() != "hello file" {
			t.Fatalf("unexpected body %q", rec.Body.String())
		}
	})

	t.Run("file", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		base := controller.NewBase(rec, req)
		if err := base.File(path); err != nil {
			t.Fatal(err)
		}
		if rec.Body.String() != "hello file" {
			t.Fatalf("unexpected body %q", rec.Body.String())
		}
	})
}

func TestRedirectRoute(t *testing.T) {
	r := router.New()
	r.Get("/posts/:id", func(w http.ResponseWriter, req *http.Request) {}, "posts.show")
	controller.SetRouter(r)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	base := controller.NewBase(rec, req)
	if err := base.RedirectRoute("posts.show", map[string]string{"id": "7"}, http.StatusSeeOther); err != nil {
		t.Fatal(err)
	}
	if loc := rec.Header().Get("Location"); loc != "/posts/7" {
		t.Fatalf("expected /posts/7, got %q", loc)
	}
}

func TestStoreUpload(t *testing.T) {
	dest := t.TempDir()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "photo.png")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("png-bytes"))
	writer.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	base := controller.NewBase(rec, req)

	saved, err := base.StoreUpload("avatar", dest, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(saved); err != nil {
		t.Fatalf("file not saved: %v", err)
	}
}

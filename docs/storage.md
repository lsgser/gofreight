# File storage

Gofreight ships a local filesystem disk for uploads, downloads, and file management — similar to Laravel's `local` disk.

## Configuration

`.env`:

```env
FILESYSTEM_DISK=local
STORAGE_LOCAL_ROOT=storage/app
```

Bootstrap (automatic in new apps):

```go
_ = app.ConfigureStorage()
// app.Storage is *storage.LocalDisk when FILESYSTEM_DISK=local
```

## LocalDisk API

```go
disk := app.Storage

// Write
_ = disk.Put(ctx, "avatars/user-1.jpg", fileReader, fileSize)

// Read
data, err := disk.Get(ctx, "avatars/user-1.jpg")

// Check existence
if disk.Exists(ctx, "avatars/user-1.jpg") { /* ... */ }

// Delete
_ = disk.Delete(ctx, "avatars/user-1.jpg")

// Filesystem path (for internal use)
absPath := disk.Path("avatars/user-1.jpg")

// Public URL path (you must serve files separately)
url := disk.URL("avatars/user-1.jpg") // → "/storage/avatars/user-1.jpg"
```

All methods accept `context.Context` as the first argument.

## Uploads in controllers

Use controller helpers (recommended):

```go
func (c PostsController) Store(base controller.Base) error {
    path, err := base.StoreUpload("attachment", "uploads/posts", 5)
    if err != nil {
        base.Unprocessable(map[string][]string{"attachment": {err.Error()}})
        return nil
    }
    // path is relative path under storage root
    post.AttachmentPath = path
    return post.Save(base.Request.Context())
}
```

| Argument | Purpose |
|----------|---------|
| `field` | Form field name (`<input name="attachment">`) |
| `destDir` | Subdirectory under storage root |
| `maxMB` | Maximum file size in megabytes |

### Download files

```go
// From disk path
return base.Download("/absolute/or/relative/path", "invoice.pdf")

// From bytes
base.DownloadBytes(pdfBytes, "report.pdf", "application/pdf")

// Stream large files
return base.StreamDownload(reader, "export.csv", "text/csv", size)
```

See [Controllers](controllers.md) for full file helper reference.

## Directory layout

```
storage/
├── app/                    ← STORAGE_LOCAL_ROOT (user uploads)
│   ├── avatars/
│   └── uploads/
├── framework/
│   ├── cache/data/         ← CACHE_STORE=file
│   └── sessions/           ← SESSION_DRIVER=file
└── logs/
```

## Serving public files

Gofreight does not auto-mount `storage/app` as public URLs. Options:

**Option A — copy to public on upload:**

```go
src, _ := app.Storage.Get(ctx, key)
_ = os.WriteFile(filepath.Join("public", "uploads", filename), src, 0644)
```

**Option B — dedicated download route:**

```go
r.Get("/files/:id", controller.Handler(c.DownloadFile))

func (c FilesController) DownloadFile(base controller.Base) error {
    id := base.Param("id")
    return base.Download(app.Storage.Path("private/"+id), "file.bin")
}
```

**Option C — static route for a public subdirectory:**

Store public-safe files under `public/uploads/` directly via `StoreUpload` targeting a public path.

## Direct disk usage

```go
func saveExport(ctx context.Context, app *application.Application, data []byte) error {
    if app.Storage == nil {
        return fmt.Errorf("storage not configured")
    }
    return app.Storage.Put(ctx, "exports/report.csv", bytes.NewReader(data), int64(len(data)))
}
```

## Validation & security

Always validate uploads in controllers:

```go
path, err := base.StoreUpload("avatar", "avatars", 2) // max 2 MB
```

Additional hardening:

- Whitelist MIME types before saving
- Generate random filenames (helpers do this by default)
- Never serve files from `storage/app` without authorization checks
- Scan uploads in production if accepting user content

See [Security](security.md).

## Testing

```go
func TestUpload(t *testing.T) {
    app := gftest.NewApp(t)
    _ = app.App.ConfigureStorage()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, _ := writer.CreateFormFile("avatar", "test.png")
    part.Write([]byte("fake-png"))
    writer.Close()

    req := httptest.NewRequest("POST", "/profile", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    // ... assert file exists on disk
}
```

## Driver matrix

| Driver | Env | Status |
|--------|-----|--------|
| Local filesystem | `FILESYSTEM_DISK=local` | **Supported** |
| S3 / cloud | — | Not built-in |
| Database BLOB | — | Use ORM |

For S3, wire a custom service implementing your own interface and register in the container.

## Related

- [Controllers](controllers.md) — `StoreUpload`, `Download`
- [Application wiring](application-wiring.md) — `ConfigureStorage`
- [Configuration](configuration.md) — env vars
- [Security](security.md) — upload hardening

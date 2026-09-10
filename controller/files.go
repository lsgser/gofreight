package controller

/*
|--------------------------------------------------------------------------
| Files
|--------------------------------------------------------------------------
|
| Implements Files as part of the controller package in the Gofreight
| framework. Key symbols: Download, File, DownloadBytes, UploadedFile,
| StoreUpload, StreamDownload.
| 
| Controllers wrap http.HandlerFunc with a Base struct that exposes
| Request, Response, validation, views, and JSON helpers.
| 
| This package defines RenderView, Redirect, status helpers, file
| downloads, and route registration utilities.
| 
| Application controllers embed these patterns; see docs/controllers.md
| for request lifecycle.
| 
| Symbols defined here include: Download (Download sends a file as an
| attachment.); File (File sends a file inline in the browser.);
| DownloadBytes (DownloadBytes sends in-memory content as a downloadable
| attachment.); UploadedFile (UploadedFile parses a multipart upload
| field.); StoreUpload (StoreUpload saves an uploaded file field to
| destDir and returns the saved path.); StreamDownload (StreamDownload
| streams a reader as a downloadable attachment.).
| 
*/

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lsgser/gofreight/upload"
)

// Download sends a file as an attachment.
func (c *Base) Download(filePath, downloadName string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("cannot download directory %q", filePath)
	}
	if downloadName == "" {
		downloadName = filepath.Base(filePath)
	}
	c.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", downloadName))
	http.ServeFile(c.Response, c.Request, filePath)
	return nil
}

// File sends a file inline in the browser.
func (c *Base) File(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("cannot serve directory %q", filePath)
	}
	http.ServeFile(c.Response, c.Request, filePath)
	return nil
}

// DownloadBytes sends in-memory content as a downloadable attachment.
func (c *Base) DownloadBytes(data []byte, downloadName, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if downloadName == "" {
		downloadName = "download"
	}
	c.Response.Header().Set("Content-Type", contentType)
	c.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", downloadName))
	c.Response.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Response.WriteHeader(http.StatusOK)
	c.Response.Write(data)
}

// UploadedFile parses a multipart upload field.
func (c *Base) UploadedFile(field string, maxMB int64) (multipart.File, *multipart.FileHeader, error) {
	return upload.ParseMultipart(c.Request, field, maxMB)
}

// StoreUpload saves an uploaded file field to destDir and returns the saved path.
func (c *Base) StoreUpload(field, destDir string, maxMB int64) (string, error) {
	file, header, err := upload.ParseMultipart(c.Request, field, maxMB)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return upload.SaveFile(file, header, destDir)
}

// StreamDownload streams a reader as a downloadable attachment.
func (c *Base) StreamDownload(r io.Reader, downloadName, contentType string, size int64) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if downloadName == "" {
		downloadName = "download"
	}
	c.Response.Header().Set("Content-Type", contentType)
	c.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", downloadName))
	if size >= 0 {
		c.Response.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	}
	c.Response.WriteHeader(http.StatusOK)
	_, err := io.Copy(c.Response, r)
	return err
}

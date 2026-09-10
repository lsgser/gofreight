package integrations

/*
|--------------------------------------------------------------------------
| Storage Client
|--------------------------------------------------------------------------
|
| Implements Storage Client as part of the integrations package in the
| Gofreight framework. Key symbols: Upload, Download, Delete.
| 
| Integrations register pluggable drivers for mail, storage, cache, queue,
| and custom third-party APIs.
| 
| Active() resolves the configured implementation from environment
| variables; wire Application in ConfigureIntegrations.
| 
| Built-in connectors cover SMTP, SendGrid, S3-compatible storage, and
| Redis without vendor-specific SDKs in app code.
| 
| Symbols defined here include: Upload (Upload puts an object into
| S3-compatible storage.); Download (Download retrieves an object from
| storage.); Delete (Delete removes an object from storage.).
| 
*/

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Upload puts an object into S3-compatible storage.
func (s *Storage) Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	client, err := s.minioClient()
	if err != nil {
		return err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err = client.PutObject(ctx, s.Bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

// Download retrieves an object from storage.
func (s *Storage) Download(ctx context.Context, key string) ([]byte, error) {
	client, err := s.minioClient()
	if err != nil {
		return nil, err
	}
	obj, err := client.GetObject(ctx, s.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

// Delete removes an object from storage.
func (s *Storage) Delete(ctx context.Context, key string) error {
	client, err := s.minioClient()
	if err != nil {
		return err
	}
	return client.RemoveObject(ctx, s.Bucket, key, minio.RemoveObjectOptions{})
}

func (s *Storage) minioClient() (*minio.Client, error) {
	if !s.enabled {
		return nil, fmt.Errorf("storage not configured")
	}
	endpoint := s.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("s3.%s.amazonaws.com", s.Region)
	}
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		Secure: true,
		Region: s.Region,
	})
}

package integrations

import (
	"fmt"
	"strings"

	"github.com/lsgser/gofreight/config"
	"github.com/redis/go-redis/v9"
)

// Storage provides S3-compatible object storage (AWS S3, MinIO, Cloudflare R2).
type Storage struct {
	Provider string
	Bucket   string
	Region   string
	Endpoint string
	enabled  bool
}

func (s *Storage) Name() string { return "storage" }

func (s *Storage) Configure(env EnvReader) error {
	disk := strings.ToLower(envGet(env, "FILESYSTEM_DISK", "STORAGE_PROVIDER"))
	if disk == "" || disk == "local" {
		s.enabled = false
		return nil
	}
	s.Provider = disk
	s.Bucket = envGet(env, "AWS_BUCKET", "STORAGE_BUCKET")
	s.Region = envGet(env, "AWS_DEFAULT_REGION", "STORAGE_REGION")
	s.Endpoint = envGet(env, "AWS_ENDPOINT", "STORAGE_ENDPOINT")
	if envGet(env, "AWS_USE_PATH_STYLE_ENDPOINT") == "true" && s.Endpoint == "" {
		s.Endpoint = envGet(env, "STORAGE_ENDPOINT")
	}
	s.enabled = s.Bucket != "" && (envGet(env, "AWS_ACCESS_KEY_ID") != "" || s.Endpoint != "")
	return nil
}

func (s *Storage) Enabled() bool { return s.enabled }

func (s *Storage) Category() Category { return CategoryStorage }

func (s *Storage) DriverID() string {
	if s.Provider != "" {
		return s.Provider
	}
	return "s3"
}

// URL returns the public URL for an object key.
func (s *Storage) URL(key string) string {
	if s.Endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.Endpoint, s.Bucket, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, key)
}

// Email provides transactional email (SMTP, SendGrid, AWS SES, Mailgun).
type Email struct {
	Provider string
	From     string
	Host     string
	APIKey   string
	enabled  bool
}

func (e *Email) Name() string { return "email" }

func (e *Email) Configure(env EnvReader) error {
	e.Provider = strings.ToLower(envGet(env, "MAIL_MAILER", "EMAIL_PROVIDER", "MAIL_DRIVER"))
	if e.Provider == "" {
		e.Provider = "smtp"
	}
	if e.Provider == "log" {
		e.enabled = false
		return nil
	}
	e.From = envGet(env, "MAIL_FROM_ADDRESS", "EMAIL_FROM")
	if name := envGet(env, "MAIL_FROM_NAME"); name != "" && e.From == "" {
		e.From = name
	}
	e.Host = envGet(env, "MAIL_HOST", "SMTP_HOST")
	e.APIKey = envGet(env, "SENDGRID_API_KEY", "EMAIL_API_KEY")
	e.enabled = e.From != "" && (e.Host != "" || e.APIKey != "" || e.Provider == "sendgrid")
	return nil
}

func (e *Email) Enabled() bool { return e.enabled }

func (e *Email) Category() Category { return CategoryMail }

func (e *Email) DriverID() string {
	if e.Provider != "" {
		return e.Provider
	}
	return "smtp"
}

// Redis provides caching and pub/sub via Redis/Upstash.
type Redis struct {
	URL     string
	enabled bool
	client  *redis.Client
}

func (r *Redis) Name() string { return "redis" }

func (r *Redis) Configure(env EnvReader) error {
	r.URL = envGet(env, "REDIS_URL", "UPSTASH_REDIS_URL")
	if r.URL == "" {
		r.URL = config.ResolveRedisURL()
	}
	r.enabled = r.URL != ""
	return nil
}

func (r *Redis) Enabled() bool { return r.enabled }

func (r *Redis) Category() Category { return CategoryCache }

func (r *Redis) DriverID() string { return "redis" }

// Webhook provides outbound webhook delivery.
type Webhook struct {
	Secret  string
	enabled bool
}

func (w *Webhook) Name() string { return "webhook" }

func (w *Webhook) Configure(env EnvReader) error {
	w.Secret = env.Get("WEBHOOK_SECRET")
	w.enabled = w.Secret != ""
	return nil
}

func (w *Webhook) Enabled() bool { return w.enabled }

func (w *Webhook) Category() Category { return CategoryWebhook }

func (w *Webhook) DriverID() string { return "webhook" }

// Analytics provides event tracking (Mixpanel, Segment, PostHog).
type Analytics struct {
	Provider string
	APIKey   string
	enabled  bool
}

func (a *Analytics) Name() string { return "analytics" }

func (a *Analytics) Configure(env EnvReader) error {
	a.Provider = env.Get("ANALYTICS_PROVIDER") // mixpanel, segment, posthog
	a.APIKey = env.Get("ANALYTICS_API_KEY")
	a.enabled = a.APIKey != ""
	return nil
}

func (a *Analytics) Enabled() bool { return a.enabled }

func (a *Analytics) Category() Category { return CategoryAnalytics }

func (a *Analytics) DriverID() string {
	if a.Provider != "" {
		return a.Provider
	}
	return "analytics"
}

// AsStorage returns the storage integration if enabled.
func AsStorage(i Integration) (*Storage, bool) {
	s, ok := i.(*Storage)
	return s, ok && s.Enabled()
}

// AsEmail returns the email integration if enabled.
func AsEmail(i Integration) (*Email, bool) {
	e, ok := i.(*Email)
	return e, ok && e.Enabled()
}

// AsRedis returns the redis integration if enabled.
func AsRedis(i Integration) (*Redis, bool) {
	r, ok := i.(*Redis)
	return r, ok && r.Enabled()
}

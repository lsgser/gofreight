package integrations

import (
	"fmt"

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
	s.Provider = env.Get("STORAGE_PROVIDER") // s3, r2, minio
	s.Bucket = env.Get("STORAGE_BUCKET")
	s.Region = env.Get("STORAGE_REGION")
	s.Endpoint = env.Get("STORAGE_ENDPOINT")
	s.enabled = s.Bucket != "" && (env.Get("AWS_ACCESS_KEY_ID") != "" || s.Endpoint != "")
	return nil
}

func (s *Storage) Enabled() bool { return s.enabled }

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
	e.Provider = env.Get("EMAIL_PROVIDER") // smtp, sendgrid, ses, mailgun
	e.From = env.Get("EMAIL_FROM")
	e.Host = env.Get("SMTP_HOST")
	e.APIKey = env.Get("EMAIL_API_KEY")
	if e.APIKey == "" {
		e.APIKey = env.Get("SENDGRID_API_KEY")
	}
	e.enabled = e.From != "" && (e.Host != "" || e.APIKey != "")
	return nil
}

func (e *Email) Enabled() bool { return e.enabled }

// Stripe provides payment processing.
type Stripe struct {
	SecretKey      string
	WebhookSecret  string
	PublishableKey string
	enabled        bool
}

func (s *Stripe) Name() string { return "stripe" }

func (s *Stripe) Configure(env EnvReader) error {
	s.SecretKey = env.Get("STRIPE_SECRET_KEY")
	s.WebhookSecret = env.Get("STRIPE_WEBHOOK_SECRET")
	s.PublishableKey = env.Get("STRIPE_PUBLISHABLE_KEY")
	s.enabled = s.SecretKey != ""
	return nil
}

func (s *Stripe) Enabled() bool { return s.enabled }

// Redis provides caching and pub/sub via Redis/Upstash.
type Redis struct {
	URL     string
	enabled bool
	client  *redis.Client
}

func (r *Redis) Name() string { return "redis" }

func (r *Redis) Configure(env EnvReader) error {
	r.URL = env.Get("REDIS_URL")
	if r.URL == "" {
		r.URL = env.Get("UPSTASH_REDIS_URL")
	}
	r.enabled = r.URL != ""
	return nil
}

func (r *Redis) Enabled() bool { return r.enabled }

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

// AsStripe returns the stripe integration if enabled.
func AsStripe(i Integration) (*Stripe, bool) {
	s, ok := i.(*Stripe)
	return s, ok && s.Enabled()
}

// AsRedis returns the redis integration if enabled.
func AsRedis(i Integration) (*Redis, bool) {
	r, ok := i.(*Redis)
	return r, ok && r.Enabled()
}

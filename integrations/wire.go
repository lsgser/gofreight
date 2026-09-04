package integrations

import (
	"os"

	"github.com/lsgser/gofreight/cache"
	"github.com/lsgser/gofreight/mail"
)

// Services holds wired application services from integrations.
type Services struct {
	Cache cache.Cacher
	Mail  mail.Mailer
}

// BuildServices creates cache and mailer from configured integrations.
func BuildServices(env EnvReader) (*Services, error) {
	if err := ConfigureAll(env); err != nil {
		return nil, err
	}

	svc := &Services{
		Cache: cache.New(),
		Mail:  mail.NewLogMailer(),
	}

	if r, ok := Get("redis"); ok {
		if redis, ok := r.(*Redis); ok && redis.Enabled() {
			if store, err := cache.NewRedis(redis.URL, "gofreight:"); err == nil {
				svc.Cache = store
			}
		}
	}

	if e, ok := Get("email"); ok {
		if em, ok := e.(*Email); ok && em.Enabled() {
			switch em.Provider {
			case "sendgrid":
				svc.Mail = mail.NewSendGrid(em.APIKey, em.From)
			default:
				svc.Mail = mail.NewSMTP(
					em.Host,
					env.Get("SMTP_PORT"),
					env.Get("SMTP_USER"),
					env.Get("SMTP_PASSWORD"),
					em.From,
				)
			}
		}
	}

	// Allow explicit mail override via env
	if os.Getenv("MAIL_DRIVER") == "log" {
		svc.Mail = mail.NewLogMailer()
	}

	return svc, nil
}

package integrations

import (
	"strings"

	"github.com/lsgser/gofreight/cache"
	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/mail"
)

// Services holds wired application services from integrations.
type Services struct {
	Cache cache.Cacher
	Mail  mail.Mailer
}

// BuildServices creates cache and mailer from configured integrations.
func BuildServices(env EnvReader) (*Services, error) {
	if err := BootProviders(env); err != nil {
		return nil, err
	}

	svc := &Services{
		Cache: cache.New(),
		Mail:  mail.NewLogMailer(),
	}

	if config.ResolveCacheStore() == "redis" {
		if redis, ok := ActiveRedis(env); ok {
			if store, err := cache.NewRedis(redis.URL, "gofreight:"); err == nil {
				svc.Cache = store
			}
		}
	}

	mailer := strings.ToLower(envGet(env, "MAIL_MAILER", "MAIL_DRIVER"))
	if mailer == "log" || mailer == "" {
		return svc, nil
	}

	if em, ok := ActiveEmail(env); ok {
		switch em.Provider {
		case "sendgrid":
			svc.Mail = mail.NewSendGrid(em.APIKey, em.From)
		default:
			svc.Mail = mail.NewSMTP(
				em.Host,
				envGet(env, "MAIL_PORT", "SMTP_PORT"),
				envGet(env, "MAIL_USERNAME", "SMTP_USER"),
				envGet(env, "MAIL_PASSWORD", "SMTP_PASSWORD"),
				em.From,
			)
		}
	}

	return svc, nil
}

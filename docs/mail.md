# Mail

Gofreight includes mailable email classes, SMTP/SendGrid delivery, and queued sending via the job system.

## Basic sending

```go
import "github.com/lsgser/gofreight/mail"

err := app.Mailer.Send(mail.Message{
    To:      []string{"user@example.com"},
    From:    "noreply@example.com",
    Subject: "Hello",
    Body:    "Plain text body",
    HTML:    "<p>HTML body</p>",
})
```

## Mailables

Mailables render HTML from template files — similar to Laravel Mailables:

```go
m := mail.NewMailable(
    "app/views/mail",       // views directory
    "welcome.html",         // template file
    "Welcome to MyApp",     // subject
    "user@example.com",     // recipient
)
m.With("Name", "Ada")
m.With("VerifyURL", verifyURL)

err := m.Send(app.Mailer)
```

Template at `app/views/mail/welcome.html`:

```html
<p>Hello, {{ .Name }}!</p>
<p><a href="{{ .VerifyURL }}">Verify your email</a></p>
```

Generate a mailable stub:

```bash
gofreight make:mail WelcomeEmail
```

Store mailable classes in `app/mail/`.

## Drivers

Configure via environment variables:

```env
# Development — logs emails instead of sending
MAIL_DRIVER=log

# SMTP
MAIL_DRIVER=smtp
MAIL_HOST=smtp.example.com
MAIL_PORT=587
MAIL_USERNAME=user
MAIL_PASSWORD=secret
MAIL_FROM=noreply@example.com

# SendGrid
MAIL_DRIVER=sendgrid
SENDGRID_API_KEY=SG.xxx
MAIL_FROM=noreply@example.com
```

`application.New()` wires the mailer from integrations. See **[Integrations](integrations.md)**.

## Queued delivery

Send email in the background so HTTP responses are not blocked:

```go
app.QueuedMailer().Send(mail.Message{
    To:      []string{"user@example.com"},
    Subject: "Your order shipped",
    HTML:    body,
})
```

Requires `app.StartJobs()` or `gofreight queue:work`. See **[Jobs & Queues](jobs.md)**.

## Authentication emails

Use mail callbacks with auth helpers:

```go
auth.RequestPasswordReset(resetStore, findUser, func(email, token string) error {
    m := mail.NewMailable("app/views/mail", "reset.html", "Reset password", email)
    m.With("ResetURL", appURL+"/password/reset?token="+token)
    return m.Send(app.Mailer)
})

auth.SendVerificationEmail(verifyStore, userID, email, func(email, token string) error {
    m := mail.NewMailable("app/views/mail", "verify.html", "Verify email", email)
    m.With("VerifyURL", appURL+"/email/verify?token="+token)
    return m.Send(app.Mailer)
})
```

See **[Authentication](authentication.md)**.

## Testing

Use `LogMailer` in tests to capture sent messages:

```go
logMailer := mail.NewLogMailer()
m.Send(logMailer)

if logMailer.Count() != 1 {
    t.Fatal("expected one email")
}
last := logMailer.Last()
```

With `gftest`:

```go
gftest.UseFakes()
gftest.AssertMailSent(t, 1)
```

## Related

- [Jobs & Queues](jobs.md) — queued mail delivery
- [Integrations](integrations.md) — SMTP, SendGrid configuration
- [Authentication](authentication.md) — password reset and verification emails

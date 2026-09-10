<!--
| --------------------------------------------------------------------------
| Mail
| --------------------------------------------------------------------------
|
| Official Gofreight documentation consumed by the docs site and linked
| from the CLI welcome page.
|
| Keep examples aligned with the current CLI and version; sync copies to
| gofreight-web when publishing.
|
| Covers concepts, tutorials, and reference material for this topic.
|

-->

<!--
| --------------------------------------------------------------------------
| Mail
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Mail

Gofreight includes mailable email classes, **GFT email templates**, SMTP/SendGrid delivery, queued sending, and a preview command for development.

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

Mailables render HTML from GFT (or legacy `.html`) templates under `app/views/mail/`:

```go
m := mail.NewMailable(
    mail.DefaultViewsRoot,      // app/views
    "mail/welcome_email_mail.gft",
    "Welcome to MyApp",
    "user@example.com",
)
m.With("Name", "Ada")
m.With("VerifyURL", verifyURL)

err := m.Send(app.Mailer)
```

Generate a mailable stub:

```bash
gofreight make:mail WelcomeEmail
```

That creates:

- `app/mail/welcome_email_mail.go` — mailable class
- `app/views/mail/welcome_email_mail.gft` — GFT content view
- `app/views/layouts/mail/default.gft` — shared email layout (if missing)

### GFT email template

`app/views/mail/welcome_email_mail.gft`:

```gft
#layout "layouts.mail.default"

#slot "content"
<p>Hello, {= .Name }!</p>
<p><a href="{= .VerifyURL }">Verify your email</a></p>
#endslot
```

Shared layout at `app/views/layouts/mail/default.gft` wraps content in a simple responsive HTML shell. Override `#layout` in individual mailables when you need a different wrapper.

Legacy `.html` templates in `app/views/mail/` still work with Go `html/template` syntax (`{{ .Name }}`).

## Preview in development

Render a mailable without sending mail:

```bash
gofreight mail:preview WelcomeEmail
gofreight mail:preview mail/welcome_email_mail --data '{"Name":"Ada"}'
gofreight mail:preview WelcomeEmail --out /tmp/welcome.html --open
```

Flags:

| Flag | Description |
|------|-------------|
| `--data JSON` | Template data (default `{}`) |
| `--out`, `-o` | Write HTML to a file |
| `--open` | Open rendered HTML in the default browser |
| `--views PATH` | Views root (default `app/views`) |

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
    m := mail.NewMailable(mail.DefaultViewsRoot, "mail/reset.html", "Reset password", email)
    m.With("ResetURL", appURL+"/password/reset?token="+token)
    return m.Send(app.Mailer)
})

auth.SendVerificationEmail(verifyStore, userID, email, func(email, token string) error {
    m := mail.NewMailable(mail.DefaultViewsRoot, "mail/verify.html", "Verify email", email)
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

Preview templates in tests with `mail.RenderView`:

```go
html, err := mail.RenderView("app/views", "mail/welcome_email_mail", map[string]any{"Name": "Test"})
```

## Related

- [Jobs & Queues](jobs.md) — queued mail delivery
- [Integrations](integrations.md) — SMTP, SendGrid configuration
- [Authentication](authentication.md) — password reset and verification emails
- [Templating](templating.md) — GFT syntax for mail views

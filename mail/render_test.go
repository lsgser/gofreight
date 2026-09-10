package mail_test

/*
|--------------------------------------------------------------------------
| Render
|--------------------------------------------------------------------------
|
| Test suite for Render in the mail package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| Mail covers Message and Mailer interfaces, LogMailer for tests,
| SMTP/SendGrid transports, mailable GFT rendering, and queued delivery.
| 
| Mailables render app/views/mail templates through the view engine;
| preview with gofreight mail:preview.
| 
| Configure MAIL_DRIVER in .env; authentication flows accept mail
| callbacks for reset and verification emails.
| 
| Run with go test ./mail/... or go test for this package from the
| framework root.
| 
*/

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lsgser/gofreight/mail"
)

func TestResolvePreviewName(t *testing.T) {
	tests := map[string]string{
		"WelcomeEmail":           "mail/welcome_email_mail",
		"WelcomeEmailMail":       "mail/welcome_email_mail",
		"welcome_email":            "mail/welcome_email_mail",
		"mail/welcome_email_mail":  "mail/welcome_email_mail",
		"mail/welcome_email_mail.gft": "mail/welcome_email_mail",
	}
	for in, want := range tests {
		if got := mail.ResolvePreviewName(in); got != want {
			t.Fatalf("ResolvePreviewName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderGFTMailWithLayout(t *testing.T) {
	dir := t.TempDir()
	layoutDir := filepath.Join(dir, "layouts", "mail")
	mailDir := filepath.Join(dir, "mail")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatal(err)
	}

	layout := `<!DOCTYPE html><html><body><div>#place "content"</div></body></html>`
	view := `#layout "layouts.mail.default"

#slot "content"
<p>Hello, {= .Name }!</p>
#endslot`

	if err := os.WriteFile(filepath.Join(layoutDir, "default.gft"), []byte(layout), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mailDir, "welcome_email.gft"), []byte(view), 0644); err != nil {
		t.Fatal(err)
	}

	html, err := mail.RenderView(dir, "mail/welcome_email", map[string]any{"Name": "Ada"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "Hello, Ada!") {
		t.Fatalf("expected rendered name, got:\n%s", html)
	}
	if strings.Contains(html, "{= .Name }") {
		t.Fatalf("GFT was not compiled: %s", html)
	}
}

func TestRenderLegacyHTMLMail(t *testing.T) {
	dir := t.TempDir()
	mailDir := filepath.Join(dir, "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatal(err)
	}
	tmpl := `<p>Hi {{ .Name }}</p>`
	if err := os.WriteFile(filepath.Join(mailDir, "notice.html"), []byte(tmpl), 0644); err != nil {
		t.Fatal(err)
	}

	html, err := mail.RenderView(dir, "mail/notice", map[string]any{"Name": "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "Hi Bob") {
		t.Fatalf("unexpected html: %s", html)
	}
}

func TestMailableLegacyViewsDir(t *testing.T) {
	dir := t.TempDir()
	mailDir := filepath.Join(dir, "mail")
	if err := os.MkdirAll(mailDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mailDir, "legacy.html"), []byte(`<b>{{ .X }}</b>`), 0644); err != nil {
		t.Fatal(err)
	}

	m := mail.NewMailable(mailDir, "legacy.html", "Test", "a@b.com")
	m.With("X", "ok")
	html, err := m.Render()
	if err != nil {
		t.Fatal(err)
	}
	if html != "<b>ok</b>" {
		t.Fatalf("got %q", html)
	}
}

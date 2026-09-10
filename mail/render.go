package mail

/*
|--------------------------------------------------------------------------
| Render
|--------------------------------------------------------------------------
|
| Implements Render as part of the mail package in the Gofreight
| framework. Key symbols: ResolveView, ResolvePreviewName, RenderView.
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
| Symbols defined here include: DefaultViewsRoot (exported value);
| ResolveView (ResolveView maps a mailable or CLI name to views root and
| template name (e.g. mail/welcome_email).); ResolvePreviewName
| (ResolvePreviewName maps CLI input (WelcomeEmail, welcome_email,
| mail/welcome_email_mail) to a template name.); RenderView (RenderView
| renders a mail template from app/views using GFT (with layouts) or
| html/template.).
| 
*/

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/lsgser/gofreight/view"
)

// DefaultViewsRoot is the standard views directory for mail templates.
const DefaultViewsRoot = "app/views"

// ResolveView maps a mailable or CLI name to views root and template name (e.g. mail/welcome_email).
func ResolveView(viewsDir, view string) (viewsRoot, templateName string) {
	view = filepath.ToSlash(view)
	viewsDir = filepath.ToSlash(viewsDir)

	if viewsDir == "" {
		viewsDir = DefaultViewsRoot
	}

	// New style: views root is app/views and view includes mail/ prefix.
	if strings.HasSuffix(viewsDir, "/views") && strings.Contains(view, "/") {
		viewsRoot = viewsDir
		templateName = templateBaseName(view)
		return viewsRoot, templateName
	}

	// Legacy: viewsDir is app/views/mail and view is a filename.
	viewsRoot = viewsDir
	if strings.HasSuffix(viewsDir, "/mail") {
		viewsRoot = filepath.ToSlash(filepath.Dir(viewsDir))
	}
	templateName = "mail/" + templateBaseName(view)
	return viewsRoot, templateName
}

// ResolvePreviewName maps CLI input (WelcomeEmail, welcome_email, mail/welcome_email_mail) to a template name.
func ResolvePreviewName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(filepath.ToSlash(name), "/")
	if strings.HasPrefix(name, "mail/") {
		return templateBaseName(name)
	}

	base := name
	if strings.HasSuffix(base, "Mail") && len(base) > 4 {
		base = base[:len(base)-4]
	}
	snake := snakeCase(base)
	if !strings.HasSuffix(snake, "_mail") {
		snake += "_mail"
	}
	return "mail/" + snake
}

func templateBaseName(path string) string {
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "/")
	for _, ext := range []string{".gft", ".html", ".tmpl", ".html.tmpl"} {
		path = strings.TrimSuffix(path, ext)
	}
	return path
}

func snakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteByte(byte(r + ('a' - 'A')))
			continue
		}
		if r == '-' || r == ' ' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

// RenderView renders a mail template from app/views using GFT (with layouts) or html/template.
func RenderView(viewsRoot, templateName string, data map[string]any) (string, error) {
	if viewsRoot == "" {
		viewsRoot = DefaultViewsRoot
	}
	if data == nil {
		data = map[string]any{}
	}

	viewPath := filepath.Join(viewsRoot, templateName)
	for _, ext := range []string{".gft", ".html", ".html.tmpl", ".tmpl"} {
		candidate := viewPath + ext
		if _, err := os.Stat(candidate); err == nil {
			return renderViewFile(viewsRoot, templateName, candidate, data)
		}
	}
	return "", fmt.Errorf("mail view not found: %s (looked under %s)", templateName, viewsRoot)
}

func renderViewFile(viewsRoot, templateName, path string, data map[string]any) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	source := string(raw)
	if strings.HasSuffix(path, view.GFTExtension) || view.IsGFTSource(source) {
		engine := view.New(viewsRoot)
		if err := engine.Load(); err != nil {
			return "", fmt.Errorf("load mail views: %w", err)
		}
		return engine.RenderString(templateName, data)
	}

	tmpl, err := template.New(filepath.Base(path)).Parse(source)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

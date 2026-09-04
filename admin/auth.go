package admin

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gofreight/gofreight/controller"
	"github.com/gofreight/gofreight/middleware"
)

// authMiddleware protects admin routes when ADMIN_PASSWORD is set.
func (p *Panel) adminPassword() string {
	if p.cfg.Password != "" {
		return p.cfg.Password
	}
	return os.Getenv("ADMIN_PASSWORD")
}

func (p *Panel) isPublicPath(path string) bool {
	return path == p.cfg.Prefix+"/login"
}

func (p *Panel) authenticated(r *http.Request) bool {
	if p.adminPassword() == "" {
		return true
	}
	session := middleware.SessionFromContext(r.Context())
	if session == nil {
		return false
	}
	authed, _ := session.Get("_admin_authed").(bool)
	return authed
}

// LoginForm shows admin login page.
func (p *Panel) LoginForm(base controller.Base) error {
	data := p.baseData(base)
	return p.render(base, "login.html", data)
}

// LoginSubmit handles admin login.
func (p *Panel) LoginSubmit(base controller.Base) error {
	password := p.adminPassword()
	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	submitted := base.Request.FormValue("password")
	if subtle.ConstantTimeCompare([]byte(submitted), []byte(password)) != 1 {
		data := p.baseData(base)
		data["Error"] = "Invalid password"
		return p.render(base, "login.html", data)
	}
	session := middleware.SessionFromContext(base.Request.Context())
	if session != nil {
		session.Set("_admin_authed", true)
	}
	base.Redirect(p.cfg.Prefix+"/", http.StatusSeeOther)
	return nil
}

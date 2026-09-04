package auth

import (
	"net/http"

	"github.com/gofreight/gofreight/controller"
	"github.com/gofreight/gofreight/middleware"
)

// LoginConfig configures session-based login.
type LoginConfig struct {
	SessionKey   string
	RedirectTo   string
	LoginPath    string
	FindUser     func(email string) (*User, error)
	AfterLogin   func(base controller.Base, user *User) error
}

// DefaultLoginConfig returns sensible login defaults.
func DefaultLoginConfig(find func(string) (*User, error)) LoginConfig {
	return LoginConfig{
		SessionKey: "current_user_id",
		RedirectTo: "/",
		LoginPath:  "/login",
		FindUser:   find,
	}
}

// Login handles POST /login form submissions.
func Login(cfg LoginConfig) func(controller.Base) error {
	return func(base controller.Base) error {
		if err := base.Request.ParseForm(); err != nil {
			return err
		}
		email := base.Request.FormValue("email")
		password := base.Request.FormValue("password")

		user, err := cfg.FindUser(email)
		if err != nil || Authenticate(user, password) != nil {
			base.Unauthorized("Invalid email or password")
			return nil
		}

		session := middleware.SessionFromContext(base.Request.Context())
		if session != nil {
			session.Set(cfg.SessionKey, user.ID)
			if user.Role != "" {
				session.Set("current_user_role", user.Role)
			}
		}

		if cfg.AfterLogin != nil {
			return cfg.AfterLogin(base, user)
		}
		base.Redirect(cfg.RedirectTo, http.StatusSeeOther)
		return nil
	}
}

// Logout clears the session and redirects.
func Logout(sessionKey, redirectTo string) func(controller.Base) error {
	return func(base controller.Base) error {
		session := middleware.SessionFromContext(base.Request.Context())
		if session != nil {
			session.Delete(sessionKey)
			session.Delete("current_user_role")
		}
		base.Redirect(redirectTo, http.StatusSeeOther)
		return nil
	}
}

// CurrentUserID returns the logged-in user ID from session.
func CurrentUserID(r *http.Request, sessionKey string) (int64, bool) {
	session := middleware.SessionFromContext(r.Context())
	if session == nil {
		return 0, false
	}
	id, ok := session.Get(sessionKey).(int64)
	if !ok {
		// JSON numbers may decode as float64
		if f, ok := session.Get(sessionKey).(float64); ok {
			return int64(f), true
		}
		return 0, false
	}
	return id, true
}

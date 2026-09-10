package auth

/*
|--------------------------------------------------------------------------
| Login
|--------------------------------------------------------------------------
|
| Implements Login as part of the auth package in the Gofreight framework.
| Key symbols: LoginConfig, DefaultLoginConfig, LoginWithJWT, Login,
| Logout, CurrentUserID.
| 
| The auth package covers session login, password hashing, API token
| storage, OAuth callbacks, email verification, and password reset flows.
| 
| Controllers compose auth helpers with your User model; tokens and
| verification stores can be in-memory or database-backed.
| 
| Install scaffolding with gofreight make:auth and wire find-user
| callbacks in app/auth.
| 
| Symbols defined here include: LoginConfig (exported type);
| DefaultLoginConfig (DefaultLoginConfig returns sensible login
| defaults.); LoginWithJWT (LoginWithJWT handles POST /login and returns a
| JWT for API clients.); Login (Login handles POST /login form
| submissions.); Logout (Logout clears the session and redirects.);
| CurrentUserID (CurrentUserID returns the logged-in user ID from
| session.).
| 
*/

import (
	"net/http"
	"time"

	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/middleware"
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

// LoginWithJWT handles POST /login and returns a JWT for API clients.
func LoginWithJWT(cfg LoginConfig, jwtMgr *JWT) func(controller.Base) error {
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

		token, expiresAt, err := jwtMgr.Issue(user.ID, user.Role)
		if err != nil {
			return err
		}

		base.RenderJSON(map[string]any{
			"token_type":   "Bearer",
			"access_token": token,
			"expires_at":   expiresAt.UTC().Format(time.RFC3339),
			"user": map[string]any{
				"id":    user.ID,
				"email": user.Email,
				"role":  user.Role,
			},
		})
		return nil
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

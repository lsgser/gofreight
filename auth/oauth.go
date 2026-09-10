package auth

/*
|--------------------------------------------------------------------------
| Oauth
|--------------------------------------------------------------------------
|
| Implements Oauth as part of the auth package in the Gofreight framework.
| Key symbols: OAuthProvider, OAuthUserInfo, OAuthConfig, OAuthRedirect,
| OAuthCallback.
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
| Symbols defined here include: OAuthProvider (exported type);
| OAuthUserInfo (exported type); OAuthConfig (exported type);
| OAuthRedirect (OAuthRedirect starts the OAuth authorization flow.);
| OAuthCallback (OAuthCallback handles the OAuth callback and logs the
| user in.).
| 
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/middleware"
)

// OAuthProvider configures a generic OAuth2 provider (no vendor lock-in).
type OAuthProvider struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	RedirectURL  string
	Scopes       []string
}

// OAuthUserInfo is normalized user info from any provider.
type OAuthUserInfo struct {
	ID    string
	Email string
	Name  string
}

// OAuthConfig wires OAuth login flow handlers.
type OAuthConfig struct {
	Provider   OAuthProvider
	OnUser     func(info OAuthUserInfo) (*User, error)
	SessionKey string
	RedirectTo string
}

// OAuthRedirect starts the OAuth authorization flow.
func OAuthRedirect(cfg OAuthConfig) func(controller.Base) error {
	return func(base controller.Base) error {
		state := randomToken()
		if session := middleware.SessionFromContext(base.Request.Context()); session != nil {
			session.Set("oauth_state_"+cfg.Provider.Name, state)
		}
		q := url.Values{
			"client_id":     {cfg.Provider.ClientID},
			"redirect_uri":  {cfg.Provider.RedirectURL},
			"response_type": {"code"},
			"scope":         {strings.Join(cfg.Provider.Scopes, " ")},
			"state":         {state},
		}
		base.Redirect(cfg.Provider.AuthURL+"?"+q.Encode(), http.StatusFound)
		return nil
	}
}

// OAuthCallback handles the OAuth callback and logs the user in.
func OAuthCallback(cfg OAuthConfig) func(controller.Base) error {
	return func(base controller.Base) error {
		state := base.Query("state")
		code := base.Query("code")
		session := middleware.SessionFromContext(base.Request.Context())
		if session != nil {
			expected, _ := session.Get("oauth_state_" + cfg.Provider.Name).(string)
			if expected == "" || expected != state {
				base.Unauthorized("Invalid OAuth state")
				return nil
			}
			session.Delete("oauth_state_" + cfg.Provider.Name)
		}
		info, err := exchangeOAuthCode(base.Request.Context(), cfg.Provider, code)
		if err != nil {
			return err
		}
		user, err := cfg.OnUser(info)
		if err != nil || user == nil {
			base.Unauthorized("OAuth login failed")
			return nil
		}
		if session != nil {
			key := cfg.SessionKey
			if key == "" {
				key = "current_user_id"
			}
			session.Set(key, user.ID)
		}
		redirect := cfg.RedirectTo
		if redirect == "" {
			redirect = "/"
		}
		base.Redirect(redirect, http.StatusSeeOther)
		return nil
	}
}

func exchangeOAuthCode(ctx context.Context, p OAuthProvider, code string) (OAuthUserInfo, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {p.RedirectURL},
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthUserInfo{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return OAuthUserInfo{}, err
	}
	defer resp.Body.Close()
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return OAuthUserInfo{}, err
	}
	return fetchOAuthUserInfo(ctx, p, tokenResp.AccessToken)
}

func fetchOAuthUserInfo(ctx context.Context, p OAuthProvider, accessToken string) (OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.UserInfoURL, nil)
	if err != nil {
		return OAuthUserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return OAuthUserInfo{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return OAuthUserInfo{}, err
	}
	return OAuthUserInfo{
		ID:    fmt.Sprint(firstKey(raw, "id", "sub")),
		Email: fmt.Sprint(firstKey(raw, "email")),
		Name:  fmt.Sprint(firstKey(raw, "name", "login")),
	}, nil
}

func firstKey(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return ""
}

func randomToken() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

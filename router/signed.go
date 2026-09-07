package router

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// URLSigner creates and verifies signed URLs.
type URLSigner struct {
	key []byte
}

// NewURLSigner creates a signer using the application key (APP_KEY).
func NewURLSigner(appKey string) *URLSigner {
	return &URLSigner{key: []byte(appKey)}
}

// Sign appends expires and signature query params to path.
func (s *URLSigner) Sign(path string, expires time.Time) string {
	u, err := url.Parse(path)
	if err != nil || u.Path == "" {
		u = &url.URL{Path: path}
	}
	q := u.Query()
	q.Set("expires", strconv.FormatInt(expires.Unix(), 10))
	u.RawQuery = q.Encode()
	unsigned := u.Path
	if u.RawQuery != "" {
		unsigned += "?" + u.RawQuery
	}
	sig := s.signature(unsigned)
	q.Set("signature", sig)
	u.RawQuery = q.Encode()
	if u.RawQuery == "" {
		return u.Path
	}
	return u.Path + "?" + u.RawQuery
}

// SignRelative signs a relative path for the given duration.
func (s *URLSigner) SignRelative(path string, ttl time.Duration) string {
	return s.Sign(path, time.Now().Add(ttl))
}

// TemporarySignedRoute builds a signed URL for a named route.
func (s *URLSigner) TemporarySignedRoute(r *Router, name string, ttl time.Duration, params map[string]string) (string, error) {
	path, err := r.URL(name, params)
	if err != nil {
		return "", err
	}
	return s.SignRelative(path, ttl), nil
}

// Verify checks whether a request URL has a valid signature and has not expired.
func (s *URLSigner) Verify(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	expires := u.Query().Get("expires")
	signature := u.Query().Get("signature")
	if expires == "" || signature == "" {
		return false
	}
	exp, err := strconv.ParseInt(expires, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}

	q := u.Query()
	q.Del("signature")
	u.RawQuery = q.Encode()
	base := u.Path
	if u.RawQuery != "" {
		base += "?" + u.RawQuery
	}
	return hmac.Equal([]byte(signature), []byte(s.signature(base)))
}

// Middleware rejects requests with missing or invalid signatures.
func (s *URLSigner) Middleware() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := r.URL.Path
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			if !s.Verify(target) {
				http.Error(w, "Invalid signature", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Signed requires a valid signed URL for this route.
func (reg *RouteRegistrar) Signed(signer *URLSigner) *RouteRegistrar {
	return reg.Use(signer.Middleware())
}

func (s *URLSigner) signature(payload string) string {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// HasValidSignature checks the current request URL.
func HasValidSignature(r *http.Request, signer *URLSigner) bool {
	if signer == nil {
		return false
	}
	target := r.URL.Path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	return signer.Verify(target)
}

// SignedPath verifies and returns the path without signature params.
func (s *URLSigner) SignedPath(rawURL string) (string, error) {
	if !s.Verify(rawURL) {
		return "", fmt.Errorf("invalid or expired signature")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Del("signature")
	q.Del("expires")
	u.RawQuery = q.Encode()
	if u.RawQuery == "" {
		return u.Path, nil
	}
	return u.Path + "?" + u.RawQuery, nil
}

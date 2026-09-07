package controller

import (
	"time"

	"github.com/lsgser/gofreight/router"
)

var defaultSigner *router.URLSigner

// SetURLSigner configures signed URL generation for controllers.
func SetURLSigner(s *router.URLSigner) {
	defaultSigner = s
}

// URLSigner returns the configured URL signer.
func URLSigner() *router.URLSigner {
	return defaultSigner
}

// Bound returns a route-model binding from the request context.
func (c *Base) Bound(name string) (any, bool) {
	return router.Bound(c.Request, name)
}

// BoundAs returns a typed route-model binding.
func BoundAs[T any](c *Base, name string) (*T, bool) {
	return router.BoundAs[T](c.Request, name)
}

// SignedURL returns a temporary signed relative URL.
func (c *Base) SignedURL(path string, ttl time.Duration) string {
	if defaultSigner == nil {
		return path
	}
	return defaultSigner.SignRelative(path, ttl)
}

// TemporarySignedRoute returns a signed URL for a named route.
func (c *Base) TemporarySignedRoute(name string, ttl time.Duration, params map[string]string) (string, error) {
	if defaultSigner == nil || defaultRouter == nil {
		return RouteURL(name, params)
	}
	return defaultSigner.TemporarySignedRoute(defaultRouter, name, ttl, params)
}

// HasValidSignature reports whether the current request URL is signed and valid.
func (c *Base) HasValidSignature() bool {
	return router.HasValidSignature(c.Request, defaultSigner)
}

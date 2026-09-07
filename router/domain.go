package router

import (
	"net/http"
	"strings"
)

// SetDefaultDomain sets the root domain used by Subdomain() groups (e.g. "example.com").
func (r *Router) SetDefaultDomain(domain string) {
	r.defaultDomain = strings.TrimSpace(domain)
}

// matchHost reports whether req.Host matches a route domain pattern.
// Patterns: "api.example.com", "{tenant}.example.com", or empty (any host).
func matchHost(reqHost, pattern string) (map[string]string, bool) {
	reqHost = stripPort(strings.ToLower(reqHost))
	if pattern == "" {
		return map[string]string{}, true
	}
	pattern = strings.ToLower(pattern)

	if !strings.Contains(pattern, "{") {
		return map[string]string{}, reqHost == pattern
	}

	reqParts := strings.Split(reqHost, ".")
	patParts := strings.Split(pattern, ".")
	if len(reqParts) != len(patParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, pp := range patParts {
		if strings.HasPrefix(pp, "{") && strings.HasSuffix(pp, "}") {
			key := pp[1 : len(pp)-1]
			params[key] = reqParts[i]
			continue
		}
		if pp != reqParts[i] {
			return nil, false
		}
	}
	return params, true
}

func stripPort(host string) string {
	if i := strings.Index(host, ":"); i >= 0 {
		return host[:i]
	}
	return host
}

func mergeHostParams(req *http.Request, params map[string]string) {
	for k, v := range params {
		req.SetPathValue("host_"+k, v)
	}
}

func resolveGroupDomain(subdomain, domain, defaultDomain string) string {
	subdomain = strings.TrimSpace(subdomain)
	domain = strings.TrimSpace(domain)

	if subdomain != "" && domain != "" {
		if strings.Contains(subdomain, "{") {
			return subdomain + "." + domain
		}
		return subdomain + "." + domain
	}
	if subdomain != "" && defaultDomain != "" {
		if strings.Contains(subdomain, "{") {
			return subdomain + "." + defaultDomain
		}
		return subdomain + "." + defaultDomain
	}
	return domain
}

// DomainFromURL extracts the host from an application URL (https://api.example.com/path → api.example.com).
func DomainFromURL(appURL string) string {
	appURL = strings.TrimSpace(appURL)
	appURL = strings.TrimPrefix(appURL, "https://")
	appURL = strings.TrimPrefix(appURL, "http://")
	if i := strings.Index(appURL, "/"); i >= 0 {
		appURL = appURL[:i]
	}
	return stripPort(appURL)
}

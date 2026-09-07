package router

import (
	"regexp"
	"strings"
)

type pathSegment struct {
	literal  string
	param    string
	optional bool
	catchAll bool
}

// normalizeRoutePath converts Laravel-style {param}, {param?}, {param*} to :param forms.
func normalizeRoutePath(path string) string {
	if !strings.Contains(path, "{") {
		return path
	}
	var b strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] != '{' {
			b.WriteByte(path[i])
			continue
		}
		end := strings.Index(path[i:], "}")
		if end < 0 {
			b.WriteString(path[i:])
			break
		}
		inner := path[i+1 : i+end]
		name := inner
		suffix := ""
		if strings.HasSuffix(inner, "?") {
			name = inner[:len(inner)-1]
			suffix = "?"
		} else if strings.HasSuffix(inner, "*") {
			name = inner[:len(inner)-1]
			suffix = "*"
		}
		b.WriteByte(':')
		b.WriteString(name)
		b.WriteString(suffix)
		i += end
	}
	return b.String()
}

func parsePathSegments(pattern string) []pathSegment {
	pattern = normalizeRoutePath(pattern)
	pattern = strings.TrimSuffix(strings.TrimSpace(pattern), "/")
	if pattern == "" || pattern == "/" {
		return nil
	}
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	segments := make([]pathSegment, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			name := part[1:]
			seg := pathSegment{param: name}
			if strings.HasSuffix(name, "?") {
				seg.param = strings.TrimSuffix(name, "?")
				seg.optional = true
			} else if strings.HasSuffix(name, "*") {
				seg.param = strings.TrimSuffix(name, "*")
				seg.catchAll = true
			}
			segments = append(segments, seg)
			continue
		}
		segments = append(segments, pathSegment{literal: part})
	}
	return segments
}

func matchPathPattern(pattern, path string) map[string]string {
	segments := parsePathSegments(pattern)
	if len(segments) == 0 {
		if path == "" || path == "/" {
			return map[string]string{}
		}
		return nil
	}

	pPath := strings.TrimSuffix(strings.TrimSpace(path), "/")
	var parts []string
	if pPath == "" || pPath == "/" {
		parts = nil
	} else {
		parts = strings.Split(strings.Trim(pPath, "/"), "/")
	}

	params := make(map[string]string)
	if matchSegments(segments, parts, 0, 0, params) {
		return params
	}
	return nil
}

func matchSegments(segments []pathSegment, parts []string, si, pi int, params map[string]string) bool {
	if si >= len(segments) {
		return pi >= len(parts)
	}

	seg := segments[si]
	if seg.catchAll {
		if pi > len(parts) {
			return false
		}
		params[seg.param] = strings.Join(parts[pi:], "/")
		return true
	}

	if seg.param != "" {
		if pi >= len(parts) {
			if seg.optional {
				return matchSegments(segments, parts, si+1, pi, params)
			}
			return false
		}
		if seg.optional {
			params[seg.param] = parts[pi]
			if matchSegments(segments, parts, si+1, pi+1, params) {
				return true
			}
			delete(params, seg.param)
			return matchSegments(segments, parts, si+1, pi, params)
		}
		params[seg.param] = parts[pi]
		return matchSegments(segments, parts, si+1, pi+1, params)
	}

	if pi >= len(parts) {
		return false
	}
	if seg.literal != parts[pi] {
		return false
	}
	return matchSegments(segments, parts, si+1, pi+1, params)
}

func routeSpecificity(path string) int {
	segments := parsePathSegments(path)
	score := 0
	for _, seg := range segments {
		switch {
		case seg.literal != "":
			score += 100
		case seg.catchAll:
			score += 1
		case seg.optional:
			score += 5
		default:
			score += 10
		}
	}
	return score
}

func compileConstraints(raw map[string]string) (map[string]*regexp.Regexp, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make(map[string]*regexp.Regexp, len(raw))
	for name, pattern := range raw {
		re, err := regexp.Compile("^" + pattern + "$")
		if err != nil {
			return nil, err
		}
		out[name] = re
	}
	return out, nil
}

func paramsMatchConstraints(params map[string]string, constraints map[string]*regexp.Regexp) bool {
	for name, re := range constraints {
		val, ok := params[name]
		if !ok || !re.MatchString(val) {
			return false
		}
	}
	return true
}

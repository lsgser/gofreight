package generator

import "strings"

func snakeCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		if r == '-' || r == ' ' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func structName(name string) string {
	parts := strings.Split(snakeCase(name), "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = title(p)
	}
	return strings.Join(parts, "")
}

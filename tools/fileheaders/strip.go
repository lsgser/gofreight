package main

/*
|--------------------------------------------------------------------------
| Strip
|--------------------------------------------------------------------------
|
| Implements Strip as part of the fileheaders package in the Gofreight
| framework.
| 
| Maintains Laravel-style file header blocks across the repository.
| 
| Run go run ./tools/fileheaders -force . from the repo root to regenerate
| extensive headers after structural changes.
| 
*/

import (
	"bytes"
	"strings"
)

func stripLeadingHeader(data []byte, ext string) []byte {
	switch ext {
	case ".go":
		return stripGoHeader(data)
	case ".gft":
		return stripGFTHeader(data)
	case ".md":
		return stripMarkdownHeader(data)
	case ".css":
		return stripCSSHeader(data)
	case ".sql", ".yaml", ".yml":
		return stripLineHeader(data)
	default:
		return data
	}
}

func stripGoHeader(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	pkgIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			pkgIdx = i
			break
		}
	}
	if pkgIdx < 0 {
		return data
	}
	start := pkgIdx + 1
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	if start >= len(lines) || strings.TrimSpace(lines[start]) != "/*" {
		return data
	}
	end := start
	for end < len(lines) {
		if strings.Contains(lines[end], "*/") {
			end++
			break
		}
		end++
	}
	for end < len(lines) && strings.TrimSpace(lines[end]) == "" {
		end++
	}
	var out []string
	out = append(out, lines[:pkgIdx+1]...)
	out = append(out, lines[end:]...)
	return []byte(strings.Join(out, "\n"))
}

func stripGFTHeader(data []byte) []byte {
	s := string(data)
	if !strings.HasPrefix(strings.TrimSpace(s), "{#") {
		return data
	}
	idx := strings.Index(s, "#}")
	if idx < 0 {
		return data
	}
	rest := strings.TrimLeft(s[idx+2:], "\n")
	return []byte(rest)
}

func stripMarkdownHeader(data []byte) []byte {
	s := strings.TrimSpace(string(data))
	if !strings.HasPrefix(s, "<!--") {
		return data
	}
	idx := strings.Index(s, "-->")
	if idx < 0 {
		return data
	}
	return []byte(strings.TrimLeft(s[idx+3:], "\n"))
}

func stripCSSHeader(data []byte) []byte {
	s := string(data)
	if !strings.HasPrefix(strings.TrimSpace(s), "/*") {
		return data
	}
	idx := strings.Index(s, "*/")
	if idx < 0 {
		return data
	}
	return bytes.TrimLeft([]byte(s[idx+2:]), "\n")
}

func stripLineHeader(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	i := 0
	for i < len(lines) {
		trim := strings.TrimSpace(lines[i])
		if trim == "" {
			i++
			continue
		}
		if strings.HasPrefix(trim, "--") || strings.HasPrefix(trim, "#") {
			if strings.Contains(lines[i], "----------") {
				i++
				continue
			}
		}
		break
	}
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i == 0 {
		return data
	}
	return []byte(strings.Join(lines[i:], "\n"))
}

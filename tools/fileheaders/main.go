package main

/*
|--------------------------------------------------------------------------
| File Headers Tool
|--------------------------------------------------------------------------
|
| Application main: loads bootstrap, registers routes, and starts the HTTP
| server.
| 
| Use gofreight serve for development; production uses a binary from
| gofreight build.
| 
| Maintains Laravel-style file header blocks across the repository.
| 
| Run go run ./tools/fileheaders -force . from the repo root to regenerate
| extensive headers after structural changes.
| 
*/

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const marker = "|--------------------------------------------------------------------------"

func main() {
	root := "."
	force := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-force", "--force":
			force = true
		default:
			root = arg
		}
	}
	var updated, skipped int
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if base == "vendor" || base == ".git" || base == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		var handler func(rel string, data []byte) ([]byte, error)
		switch ext {
		case ".go":
			handler = insertGoHeader
		case ".gft":
			handler = insertGFTHeader
		case ".css":
			handler = insertCSSHeader
		case ".sql":
			handler = insertSQLHeader
		case ".yaml", ".yml":
			handler = insertYAMLHeader
		case ".md":
			handler = insertMarkdownHeader
		default:
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if hasLeadingHeader(data) && !force {
			skipped++
			return nil
		}
		if force && hasLeadingHeader(data) {
			data = stripLeadingHeader(data, ext)
		}
		newContent, err := handler(rel, data)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		if err := os.WriteFile(path, newContent, 0644); err != nil {
			return err
		}
		updated++
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Updated %d files, skipped %d (use -force to replace existing headers).\n", updated, skipped)
}

func hasLeadingHeader(data []byte) bool {
	lines := strings.Split(string(data), "\n")
	end := 40
	if end > len(lines) {
		end = len(lines)
	}
	for i := 0; i < end; i++ {
		if strings.Contains(lines[i], marker) {
			return true
		}
	}
	return false
}

func insertGoHeader(rel string, data []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, rel, data, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	header := buildGoHeader(rel, f)
	lines := strings.Split(string(data), "\n")
	pkgIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			pkgIdx = i
			break
		}
	}
	if pkgIdx < 0 {
		return nil, fmt.Errorf("no package clause")
	}
	headerLines := strings.Split(strings.TrimRight(header, "\n"), "\n")
	var out []string
	out = append(out, lines[:pkgIdx+1]...)
	out = append(out, "")
	out = append(out, headerLines...)
	out = append(out, "")
	start := pkgIdx + 1
	if start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	out = append(out, lines[start:]...)
	return []byte(strings.Join(out, "\n")), nil
}

func buildGoHeader(rel string, f *ast.File) string {
	title := fileTitle(rel)
	paragraphs := extensiveParagraphs(rel, f)
	var b strings.Builder
	b.WriteString("/*\n")
	b.WriteString("|--------------------------------------------------------------------------\n")
	b.WriteString("| " + title + "\n")
	b.WriteString("|--------------------------------------------------------------------------\n")
	b.WriteString("|\n")
	writeParagraphs(&b, "| ", paragraphs)
	b.WriteString("*/\n")
	return b.String()
}

func writeParagraphs(b *strings.Builder, prefix string, paragraphs []string) {
	for _, p := range paragraphs {
		for _, line := range wrapLine(p, 72) {
			b.WriteString(prefix + line + "\n")
		}
		b.WriteString(prefix + "\n")
	}
}

func fileTitle(rel string) string {
	base := filepath.Base(rel)
	base = strings.TrimSuffix(base, ".go")
	isTest := strings.HasSuffix(rel, "_test.go")
	if isTest {
		base = strings.TrimSuffix(base, "_test")
	}
	if base == "main" {
		dir := filepath.ToSlash(filepath.Dir(rel))
		switch {
		case dir == "cmd/gofreight":
			return "Gofreight CLI Entry Point"
		case strings.HasSuffix(dir, "tools/migrate"):
			return "Migration Tool Entry Point"
		case dir == "tools/fileheaders":
			return "File Headers Tool"
		case dir == "demoapp" || dir == "examples/blog":
			return "Application Entry Point"
		default:
			return "Application Entry Point"
		}
	}
	if base == "doc" {
		return humanPackage(filepath.Dir(rel)) + " Overview"
	}
	return humanIdent(base)
}

func humanPackage(dir string) string {
	dir = strings.ReplaceAll(dir, "/", " — ")
	parts := strings.Split(dir, " — ")
	for i, p := range parts {
		parts[i] = humanIdent(p)
	}
	return strings.Join(parts, " — ")
}

func humanIdent(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		if r[i-1] == '_' {
			r[i] = unicode.ToUpper(r[i])
		}
	}
	out := strings.ReplaceAll(string(r), "_", " ")
	out = strings.ReplaceAll(out, "  ", " ")
	return out
}

func insertGFTHeader(rel string, data []byte) ([]byte, error) {
	title, paragraphs := extensiveTextMeta(rel)
	block := formatBlock(title, paragraphs, "| ", "|\n", "{#\n", "#}\n\n")
	return append([]byte(block), data...), nil
}

func insertCSSHeader(rel string, data []byte) ([]byte, error) {
	title, paragraphs := extensiveTextMeta(rel)
	block := formatBlock(title, paragraphs, " * ", " *\n", "/*\n", "\n */")
	return append(append([]byte(block), '\n'), data...), nil
}

func insertSQLHeader(rel string, data []byte) ([]byte, error) {
	title, paragraphs := extensiveTextMeta(rel)
	block := formatBlock(title, paragraphs, "-- ", "--\n", "", "")
	return append([]byte(block), data...), nil
}

func insertYAMLHeader(rel string, data []byte) ([]byte, error) {
	title, paragraphs := extensiveTextMeta(rel)
	block := formatBlock(title, paragraphs, "# ", "#\n", "", "")
	return append([]byte(block), data...), nil
}

func insertMarkdownHeader(rel string, data []byte) ([]byte, error) {
	title, paragraphs := extensiveTextMeta(rel)
	if strings.HasPrefix(string(data), "# ") {
		first := strings.SplitN(string(data), "\n", 2)[0]
		title = strings.TrimPrefix(first, "# ")
	}
	block := formatBlock(title, paragraphs, "| ", "|\n", "<!--\n", "\n-->\n\n")
	return append([]byte(block), data...), nil
}

func formatBlock(title string, paragraphs []string, prefix, blankLine, open, close string) string {
	var b strings.Builder
	b.WriteString(open)
	b.WriteString(prefix + strings.TrimPrefix(marker, "|") + "\n")
	b.WriteString(prefix + title + "\n")
	b.WriteString(prefix + strings.TrimPrefix(marker, "|") + "\n")
	b.WriteString(blankLine)
	for _, p := range paragraphs {
		for _, line := range wrapLine(p, 72) {
			b.WriteString(prefix + line + "\n")
		}
		b.WriteString(blankLine)
	}
	b.WriteString(close)
	if !strings.HasSuffix(close, "\n") && close != "" {
		b.WriteString("\n")
	}
	return b.String()
}

func wrapLine(s string, width int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	var cur strings.Builder
	for _, w := range words {
		if cur.Len() == 0 {
			cur.WriteString(w)
			continue
		}
		if cur.Len()+1+len(w) > width {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
			continue
		}
		cur.WriteString(" ")
		cur.WriteString(w)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

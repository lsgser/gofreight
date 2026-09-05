package view

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Engine renders HTML templates with layouts and partials using Gofreight Templates (GFT).
type Engine struct {
	mu        sync.RWMutex
	templates *template.Template
	layout    string
	funcs     template.FuncMap
	root      string
	meta      map[string]CompiledGFT
}

// New creates a view engine rooted at the given directory (e.g. "app/views").
func New(root string) *Engine {
	return &Engine{
		root:   root,
		layout: "layouts/application",
		funcs:  defaultFuncs(),
		meta:   make(map[string]CompiledGFT),
	}
}

// RegisterFunc adds a template helper function.
func (e *Engine) RegisterFunc(name string, fn any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcs[name] = fn
	e.templates = nil
}

// SetLayout sets the default layout template name.
func (e *Engine) SetLayout(name string) {
	e.layout = name
}

// Load parses view files under the views root (.gft and .html with GFT syntax).
func (e *Engine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tmpl := template.New("gofreight").Funcs(e.funcs)
	meta := make(map[string]CompiledGFT)

	err := filepath.WalkDir(e.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isTemplateFile(path) {
			return err
		}
		name := templateName(e.root, path)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(content)
		if strings.HasSuffix(path, GFTExtension) || IsGFTSource(source) {
			compiled, err := CompileGFT(source)
			if err != nil {
				return fmt.Errorf("compile gft %s: %w", name, err)
			}
			meta[name] = compiled
			source = compiled.Source
		}
		_, err = tmpl.New(name).Parse(source)
		return err
	})
	if err != nil {
		return fmt.Errorf("load views: %w", err)
	}
	e.templates = tmpl
	e.meta = meta
	return nil
}

// Render executes a template with optional layout wrapping.
func (e *Engine) Render(w http.ResponseWriter, name string, data any) error {
	layout := e.layout
	if m, ok := e.meta[name]; ok && m.Layout != "" {
		layout = m.Layout
	}
	return e.RenderWithLayout(w, name, layout, data)
}

// RenderWithLayout renders a template wrapped in the specified layout.
func (e *Engine) RenderWithLayout(w http.ResponseWriter, name, layout string, data any) error {
	tmpl, err := e.getTemplates()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	meta := e.meta[name]
	if meta.Layout != "" {
		layout = meta.Layout
	}
	if layout == "" {
		return tmpl.ExecuteTemplate(w, name, data)
	}

	content := ""
	if sec := meta.Sections["content"]; sec != "" {
		content, err = e.executeSource(sec, data)
		if err != nil {
			return err
		}
	} else {
		content, err = e.executeString(tmpl, name, data)
		if err != nil {
			return err
		}
	}

	layoutData := mergeLayoutData(data, content, meta)
	return tmpl.ExecuteTemplate(w, layout, layoutData)
}

func mergeLayoutData(data any, content string, meta CompiledGFT) map[string]any {
	layoutData := map[string]any{
		"Content": template.HTML(content),
		"Data":    data,
		"Sections": map[string]template.HTML{},
	}
	if m, ok := data.(map[string]any); ok {
		for k, v := range m {
			layoutData[k] = v
		}
	}
	sections := layoutData["Sections"].(map[string]template.HTML)
	for k, v := range meta.Sections {
		sections[k] = template.HTML(v)
	}
	if _, ok := sections["content"]; !ok && content != "" {
		sections["content"] = template.HTML(content)
		layoutData["Content"] = template.HTML(content)
	}
	layoutData["Sections"] = sections
	return layoutData
}

// Partial renders a partial template (no layout).
func (e *Engine) Partial(w http.ResponseWriter, name string, data any) error {
	tmpl, err := e.getTemplates()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, name, data)
}

func (e *Engine) getTemplates() (*template.Template, error) {
	e.mu.RLock()
	tmpl := e.templates
	e.mu.RUnlock()

	if tmpl != nil {
		return tmpl, nil
	}
	if err := e.Load(); err != nil {
		return nil, err
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.templates, nil
}

func (e *Engine) executeString(tmpl *template.Template, name string, data any) (string, error) {
	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (e *Engine) executeSource(source string, data any) (string, error) {
	t, err := template.New("inline").Funcs(e.funcs).Parse(source)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func defaultFuncs() template.FuncMap {
	return template.FuncMap{
		"upper":       strings.ToUpper,
		"lower":       strings.ToLower,
		"title":       title,
		"join":        strings.Join,
		"contains":    strings.Contains,
		"trim":        strings.TrimSpace,
		"default":     defaultVal,
		"safeHTML":    func(s string) template.HTML { return template.HTML(s) },
		"safeURL":     func(s string) template.URL { return template.URL(s) },
		"old":         oldField,
		"fieldErrors": fieldErrors,
		"hasError":    hasFieldError,
	}
}

func oldField(field string, ctx map[string]any, fallback string) string {
	return Old(field, ctx, fallback)
}

func fieldErrors(field string, ctx map[string]any) []string {
	return FieldErrors(field, ctx)
}

func hasFieldError(field string, ctx map[string]any) bool {
	return HasError(field, ctx)
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func defaultVal(def, val string) string {
	if val == "" {
		return def
	}
	return val
}

func isTemplateFile(path string) bool {
	return strings.HasSuffix(path, GFTExtension) ||
		strings.HasSuffix(path, ".html") ||
		strings.HasSuffix(path, ".html.tmpl") ||
		strings.HasSuffix(path, ".tmpl")
}

func templateName(root, path string) string {
	rel, _ := filepath.Rel(root, path)
	name := strings.TrimSuffix(rel, GFTExtension)
	name = strings.TrimSuffix(name, ".tmpl")
	name = strings.TrimSuffix(name, ".html")
	return filepath.ToSlash(name)
}

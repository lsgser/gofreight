package view

import (
	"fmt"
	"regexp"
	"strings"
)

// GFT (Gofreight Template) is Gofreight's view language — inspired by Laravel and
// Rails conventions but with its own syntax and file extension (.gft).

// CompiledGFT holds the result of compiling a .gft template.
type CompiledGFT struct {
	Layout   string
	Sections map[string]string
	Source   string
}

var (
	reGFTComment    = regexp.MustCompile(`(?s)\{#.*?#\}`)
	reGFTOutput     = regexp.MustCompile(`\{=\s*(.+?)\s*\}`)
	reGFTRawOutput  = regexp.MustCompile(`\{!\s*(.+?)\s*!\}`)
	reGFTLayout     = regexp.MustCompile(`(?m)^#layout\s+["']([^"']+)["']\s*$`)
	reGFTSlot       = regexp.MustCompile(`(?s)#slot\s+["']([^"']+)["'](.*?)#endslot`)
	reGFTPartial    = regexp.MustCompile(`#partial\s+["']([^"']+)["']`)
	reGFTEachOr     = regexp.MustCompile(`(?s)#eachor\s+([^\n]+)\n(.*?)#otherwise(.*?)#endeach`)
	reGFTEach       = regexp.MustCompile(`#each\s+(.+?)\s+as\s+(\w+)`)
	reGFTEndEach    = regexp.MustCompile(`#endeach\b`)
	reGFTWhen       = regexp.MustCompile(`#when\s+(.+)`)
	reGFTOtherwise  = regexp.MustCompile(`#otherwise\b`)
	reGFTEndWhen    = regexp.MustCompile(`#endwhen\b`)
	reGFTOrWhen     = regexp.MustCompile(`#orwhen\s+(.+)`)
	reGFTUnless     = regexp.MustCompile(`#unless\s+(.+)`)
	reGFTEndUnless  = regexp.MustCompile(`#endunless\b`)
	reGFTSignedIn   = regexp.MustCompile(`#signedin\b`)
	reGFTSignedOut  = regexp.MustCompile(`#signedout\b`)
	reGFTEndSigned  = regexp.MustCompile(`#endsigned(?:in|out)\b`)
	reGFTToken      = regexp.MustCompile(`#token\b`)
	reGFTForm       = regexp.MustCompile(`(?s)#form\s+([^\n]+)\n(.*?)#endform`)
	reGFTField      = regexp.MustCompile(`(?m)^#field\s+(.+)$`)
	reGFTError      = regexp.MustCompile(`#error\s+["']([^"']+)["']`)
	reGFTPlace      = regexp.MustCompile(`#place\s+["']([^"']+)["'](?:\s+["']([^"']*)["'])?`)
)

// CompileGFT transforms Gofreight Template syntax into Go html/template.
func CompileGFT(input string) (CompiledGFT, error) {
	result := CompiledGFT{Sections: make(map[string]string)}

	if m := reGFTLayout.FindStringSubmatch(input); len(m) > 1 {
		result.Layout = gftPathToName(m[1])
		input = reGFTLayout.ReplaceAllString(input, "")
	}

	for _, m := range reGFTSlot.FindAllStringSubmatch(input, -1) {
		name := m[1]
		body := strings.TrimSpace(m[2])
		compiled, err := compileGFTBody(body)
		if err != nil {
			return result, err
		}
		result.Sections[name] = compiled
		input = strings.Replace(input, m[0], "", 1)
	}

	body, err := compileGFTBody(strings.TrimSpace(input))
	if err != nil {
		return result, err
	}
	result.Source = body
	return result, nil
}

func compileGFTBody(input string) (string, error) {
	s := input
	s = reGFTComment.ReplaceAllString(s, "")

	// {= expr } → Go template output
	s = reGFTOutput.ReplaceAllStringFunc(s, func(m string) string {
		parts := reGFTOutput.FindStringSubmatch(m)
		if len(parts) < 2 {
			return m
		}
		return gftOutputExpr(parts[1])
	})
	s = reGFTRawOutput.ReplaceAllStringFunc(s, func(m string) string {
		parts := reGFTRawOutput.FindStringSubmatch(m)
		if len(parts) < 2 {
			return m
		}
		inner := strings.TrimSpace(parts[1])
		if strings.HasPrefix(inner, ".") {
			return fmt.Sprintf("{{safeHTML %s}}", inner)
		}
		return fmt.Sprintf("{{safeHTML %s}}", inner)
	})

	// #eachor ... #otherwise ... #endeach
	s = reGFTEachOr.ReplaceAllStringFunc(s, func(match string) string {
		parts := reGFTEachOr.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		open, _ := eachHeader(parts[1])
		return open + parts[2] + `{{else}}` + parts[3] + `{{end}}`
	})

	s = reGFTEach.ReplaceAllStringFunc(s, func(m string) string {
		parts := reGFTEach.FindStringSubmatch(m)
		if len(parts) < 3 {
			return m
		}
		return fmt.Sprintf(`{{range $%s := %s}}`, parts[2], normalizeGFTExpr(parts[1]))
	})
	s = reGFTEndEach.ReplaceAllString(s, `{{end}}`)

	s = reGFTUnless.ReplaceAllString(s, `{{if not $1}}`)
	s = reGFTEndUnless.ReplaceAllString(s, `{{end}}`)
	s = reGFTWhen.ReplaceAllString(s, `{{if $1}}`)
	s = reGFTOrWhen.ReplaceAllString(s, `{{else if $1}}`)
	s = reGFTOtherwise.ReplaceAllString(s, `{{else}}`)
	s = reGFTEndWhen.ReplaceAllString(s, `{{end}}`)

	s = reGFTSignedIn.ReplaceAllString(s, `{{if .CurrentUser}}`)
	s = reGFTEndSigned.ReplaceAllString(s, `{{end}}`)
	s = reGFTSignedOut.ReplaceAllString(s, `{{if not .CurrentUser}}`)

	s = reGFTToken.ReplaceAllString(s, `<input type="hidden" name="authenticity_token" value="{{.CSRFToken}}">`)

	s = reGFTForm.ReplaceAllStringFunc(s, compileGFTForm)
	s = reGFTField.ReplaceAllStringFunc(s, compileGFTField)
	s = reGFTError.ReplaceAllStringFunc(s, compileGFTError)

	s = reGFTPartial.ReplaceAllString(s, `{{template "$1" .}}`)

	s = reGFTPlace.ReplaceAllStringFunc(s, func(m string) string {
		parts := reGFTPlace.FindStringSubmatch(m)
		if len(parts) < 2 {
			return m
		}
		if parts[1] == "content" {
			return "{{.Content}}"
		}
		def := ""
		if len(parts) > 2 {
			def = parts[2]
		}
		return fmt.Sprintf(`{{with index .Sections "%s"}}{{.}}{{else}}%s{{end}}`, parts[1], def)
	})

	return s, nil
}

func eachHeader(header string) (string, error) {
	header = strings.TrimSpace(header)
	if strings.Contains(header, " as ") {
		parts := strings.SplitN(header, " as ", 2)
		collection := normalizeGFTExpr(strings.TrimSpace(parts[0]))
		v := strings.TrimSpace(parts[1])
		return fmt.Sprintf(`{{range $%s := %s}}`, v, collection), nil
	}
	return fmt.Sprintf(`{{range %s}}`, normalizeGFTExpr(header)), nil
}

func gftOutputExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, ".") {
		return fmt.Sprintf("{{%s}}", expr)
	}
	return fmt.Sprintf("{{ %s }}", expr)
}

func normalizeGFTExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, ".") {
		return expr
	}
	// bare name → .Name (PascalCase field on data map/struct)
	if len(expr) > 0 && expr[0] >= 'a' && expr[0] <= 'z' {
		return "." + strings.ToUpper(expr[:1]) + expr[1:]
	}
	return expr
}

func gftPathToName(path string) string {
	path = strings.Trim(path, "'\"")
	return strings.ReplaceAll(path, ".", "/")
}

// IsGFTSource returns true if the source uses Gofreight Template directives.
func IsGFTSource(s string) bool {
	markers := []string{
		"#layout", "#slot", "#endslot", "#partial", "#each", "#endeach",
		"#eachor", "#when", "#endwhen", "#place", "#token", "#form", "#endform",
		"#field", "#error", "{=", "{!",
	}
	for _, m := range markers {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}

// GFTExtension is the recommended view file extension.
const GFTExtension = ".gft"

func compileGFTForm(match string) string {
	parts := reGFTForm.FindStringSubmatch(match)
	if len(parts) < 3 {
		return match
	}
	attrs := parseGFTAttrs(parts[1])
	action := attrs["action"]
	method := strings.ToUpper(defaultStr(attrs["method"], "POST"))
	body := parts[2]

	var b strings.Builder
	b.WriteString(`<form method="POST" action="`)
	b.WriteString(action)
	b.WriteString(`">`)
	b.WriteString("\n  ")
	b.WriteString(`<input type="hidden" name="authenticity_token" value="{{.CSRFToken}}">`)
	if method != "POST" && method != "GET" {
		b.WriteString("\n  ")
		b.WriteString(`<input type="hidden" name="_method" value="`)
		b.WriteString(method)
		b.WriteString(`">`)
	}
	b.WriteString("\n")
	b.WriteString(body)
	b.WriteString("\n</form>")
	return b.String()
}

func compileGFTField(match string) string {
	parts := reGFTField.FindStringSubmatch(match)
	if len(parts) < 2 {
		return match
	}
	attrs := parseGFTAttrs(parts[1])
	name := attrs["name"]
	label := defaultStr(attrs["label"], titleCase(name))
	fieldType := defaultStr(attrs["type"], "text")
	valueExpr := attrs["value"]
	if valueExpr == "" {
		valueExpr = ".Item." + titleCase(name)
	}

	var b strings.Builder
	b.WriteString(`<div class="field{{if hasError "`)
	b.WriteString(name)
	b.WriteString(`" .}} field-error{{end}}">`)
	b.WriteString("\n  ")
	b.WriteString(`<label for="`)
	b.WriteString(name)
	b.WriteString(`">`)
	b.WriteString(label)
	b.WriteString(`</label>`)
	b.WriteString("\n  ")

	switch fieldType {
	case "textarea":
		b.WriteString(`<textarea name="`)
		b.WriteString(name)
		b.WriteString(`" id="`)
		b.WriteString(name)
		b.WriteString(`">{{ old "`)
		b.WriteString(name)
		b.WriteString(`" . (printf "%v" `)
		b.WriteString(valueExpr)
		b.WriteString(`) }}</textarea>`)
	case "select":
		b.WriteString(`<select name="`)
		b.WriteString(name)
		b.WriteString(`" id="`)
		b.WriteString(name)
		b.WriteString(`">`)
		for _, opt := range strings.Split(attrs["options"], ",") {
			opt = strings.TrimSpace(opt)
			if opt == "" {
				continue
			}
			b.WriteString("\n    <option value=\"")
			b.WriteString(opt)
			b.WriteString(`">{{ title "`)
			b.WriteString(opt)
			b.WriteString(`" }}</option>`)
		}
		b.WriteString("\n  </select>")
	case "checkbox":
		b.WriteString(`<label><input type="checkbox" name="`)
		b.WriteString(name)
		b.WriteString(`" value="1"> `)
		b.WriteString(label)
		b.WriteString(`</label>`)
	default:
		b.WriteString(`<input type="`)
		b.WriteString(fieldType)
		b.WriteString(`" name="`)
		b.WriteString(name)
		b.WriteString(`" id="`)
		b.WriteString(name)
		b.WriteString(`" value="{{ old "`)
		b.WriteString(name)
		b.WriteString(`" . (printf "%v" `)
		b.WriteString(valueExpr)
		b.WriteString(`) }}">`)
	}

	if fieldType != "checkbox" {
		b.WriteString("\n  ")
		b.WriteString(`{{range fieldErrors "`)
		b.WriteString(name)
		b.WriteString(`" .}}<p class="field-error">{{.}}</p>{{end}}`)
	}
	b.WriteString("\n</div>")
	return b.String()
}

func compileGFTError(match string) string {
	parts := reGFTError.FindStringSubmatch(match)
	if len(parts) < 2 {
		return match
	}
	field := parts[1]
	return fmt.Sprintf(`{{range fieldErrors "%s" .}}<p class="field-error">{{.}}</p>{{end}}`, field)
}

func parseGFTAttrs(line string) map[string]string {
	attrs := make(map[string]string)
	re := regexp.MustCompile(`(?:(\w+)=)?["']([^"']*)["']`)
	for _, m := range re.FindAllStringSubmatch(line, -1) {
		if len(m) < 3 {
			continue
		}
		key := m[1]
		val := m[2]
		if key == "" {
			if attrs["name"] == "" {
				attrs["name"] = val
			}
			continue
		}
		attrs[key] = val
	}
	return attrs
}

func defaultStr(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}


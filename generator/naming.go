package generator

import (
	"fmt"
	"strings"
)

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


// ParsedField holds generator mappings for a model/scaffold field.
type ParsedField struct {
	Name           string
	DBTag          string
	JSONTag        string
	RawType        string
	GoType         string
	SQLType        string
	FormType       string
	EnumValues     []string
	ReferenceTable string
}

// ParseField maps a CLI field spec (e.g. "title:string", "status:enum:draft,published") to Go/SQL/form types.
func ParseField(name, typeSpec string) ParsedField {
	typeSpec = strings.TrimSpace(strings.ToLower(typeSpec))
	base, extra := splitTypeSpec(typeSpec)

	pf := ParsedField{
		Name:    title(name),
		DBTag:   strings.ToLower(name),
		RawType: typeSpec,
	}
	pf.JSONTag = pf.DBTag

	switch base {
	case "string", "str":
		pf.GoType, pf.SQLType, pf.FormType = "string", "VARCHAR(255) NOT NULL", "text"
	case "text":
		pf.GoType, pf.SQLType, pf.FormType = "string", "TEXT NOT NULL", "textarea"
	case "email":
		pf.GoType, pf.SQLType, pf.FormType = "string", "VARCHAR(255) NOT NULL", "email"
	case "url":
		pf.GoType, pf.SQLType, pf.FormType = "string", "VARCHAR(512) NOT NULL", "url"
	case "int", "integer":
		pf.GoType, pf.SQLType, pf.FormType = "int", "INTEGER NOT NULL", "number"
	case "bigint":
		pf.GoType, pf.SQLType, pf.FormType = "int64", "INTEGER NOT NULL", "number"
	case "float", "decimal", "double":
		pf.GoType, pf.SQLType, pf.FormType = "float64", "REAL NOT NULL", "number"
	case "bool", "boolean":
		pf.GoType, pf.SQLType, pf.FormType = "bool", "INTEGER NOT NULL DEFAULT 0", "checkbox"
	case "datetime", "timestamp":
		pf.GoType, pf.SQLType, pf.FormType = "string", "TEXT", "datetime"
	case "date":
		pf.GoType, pf.SQLType, pf.FormType = "string", "TEXT", "date"
	case "time":
		pf.GoType, pf.SQLType, pf.FormType = "string", "TEXT", "time"
	case "uuid":
		pf.GoType, pf.SQLType, pf.FormType = "string", "TEXT NOT NULL", "text"
	case "json", "jsonb":
		pf.GoType, pf.SQLType, pf.FormType = "string", "TEXT NOT NULL", "textarea"
	case "enum":
		pf.GoType = "string"
		pf.EnumValues = parseEnumValues(extra)
		pf.SQLType = "TEXT NOT NULL"
		pf.FormType = "select"
	case "references", "reference", "belongs_to":
		pf.GoType = "int64"
		pf.SQLType = "INTEGER NOT NULL"
		pf.FormType = "number"
		pf.ReferenceTable = extra
	default:
		pf.GoType, pf.SQLType, pf.FormType = "string", "VARCHAR(255) NOT NULL", "text"
	}

	return pf
}

// FakerExpr returns a Go expression for fake data in generated factories.
func (pf ParsedField) FakerExpr() string {
	base, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(pf.RawType)), ":")
	switch base {
	case "email":
		return "gfaker.Email()"
	case "url":
		return "gfaker.URL()"
	case "text":
		return "gfaker.Paragraph()"
	case "int", "integer":
		return "gfaker.Int()"
	case "bigint", "references", "reference", "belongs_to":
		return "gfaker.Int64()"
	case "float", "decimal", "double":
		return "gfaker.Float()"
	case "bool", "boolean":
		return "gfaker.Bool()"
	case "datetime", "timestamp":
		return "gfaker.DateTime()"
	case "date":
		return "gfaker.Date()"
	case "time":
		return "gfaker.Time()"
	case "uuid":
		return "gfaker.UUID()"
	case "json", "jsonb":
		return `"{}"`
	case "enum":
		if len(pf.EnumValues) > 0 {
			return `"` + pf.EnumValues[0] + `"`
		}
		return "gfaker.Word()"
	case "string", "str":
		if strings.Contains(pf.DBTag, "email") {
			return "gfaker.Email()"
		}
		if strings.Contains(pf.DBTag, "title") || strings.Contains(pf.DBTag, "subject") {
			return "gfaker.Sentence()"
		}
		if strings.Contains(pf.DBTag, "name") {
			return "gfaker.Name()"
		}
		return "gfaker.Word()"
	default:
		return "gfaker.Word()"
	}
}

func splitTypeSpec(spec string) (base, extra string) {
	switch {
	case strings.HasPrefix(spec, "enum:"):
		return "enum", strings.TrimPrefix(spec, "enum:")
	case strings.HasPrefix(spec, "references:"):
		return "references", strings.TrimPrefix(spec, "references:")
	case strings.HasPrefix(spec, "reference:"):
		return "references", strings.TrimPrefix(spec, "reference:")
	case strings.HasPrefix(spec, "belongs_to:"):
		return "belongs_to", strings.TrimPrefix(spec, "belongs_to:")
	}
	parts := strings.SplitN(spec, ":", 2)
	base = parts[0]
	if len(parts) > 1 {
		extra = parts[1]
	}
	return base, extra
}

func parseEnumValues(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// MigrationColumnDef returns the SQL column definition for migrations.
func (pf ParsedField) MigrationColumnDef() string {
	if len(pf.EnumValues) > 0 {
		quoted := make([]string, len(pf.EnumValues))
		for i, v := range pf.EnumValues {
			quoted[i] = "'" + strings.ReplaceAll(v, "'", "''") + "'"
		}
		return fmt.Sprintf("TEXT NOT NULL CHECK (%s IN (%s))", pf.DBTag, strings.Join(quoted, ", "))
	}
	return pf.SQLType
}

// HTMLInputType returns the HTML input type for scaffold forms.
func (pf ParsedField) HTMLInputType() string {
	if pf.FormType == "datetime" {
		return "datetime-local"
	}
	if pf.FormType == "text" || pf.FormType == "textarea" || pf.FormType == "select" || pf.FormType == "checkbox" {
		return pf.FormType
	}
	return pf.FormType
}

// ScaffoldFieldTypes lists supported generator field types for documentation.
func ScaffoldFieldTypes() []struct {
	Type, Aliases, GoType, SQLType, Form, Example string
} {
	return []struct {
		Type, Aliases, GoType, SQLType, Form, Example string
	}{
		{"string", "str", "string", "VARCHAR(255)", "text input", "title:string"},
		{"text", "", "string", "TEXT", "textarea", "body:text"},
		{"email", "", "string", "VARCHAR(255)", "email input", "contact:email"},
		{"url", "", "string", "VARCHAR(512)", "url input", "website:url"},
		{"integer", "int", "int", "INTEGER", "number input", "views:integer"},
		{"bigint", "", "int64", "INTEGER", "number input", "legacy_id:bigint"},
		{"float", "decimal, double", "float64", "REAL", "number input", "price:float"},
		{"boolean", "bool", "bool", "INTEGER (0/1)", "checkbox", "published:boolean"},
		{"datetime", "timestamp", "string", "TEXT", "datetime-local", "published_at:datetime"},
		{"date", "", "string", "TEXT", "date input", "starts_on:date"},
		{"time", "", "string", "TEXT", "time input", "opens_at:time"},
		{"uuid", "", "string", "TEXT", "text input", "token:uuid"},
		{"json", "jsonb", "string", "TEXT", "textarea", "metadata:json"},
		{"enum", "", "string", "TEXT + CHECK", "select", "status:enum:draft,published,archived"},
		{"references", "reference, belongs_to", "int64", "INTEGER", "number (FK id)", "post_id:references:posts"},
	}
}

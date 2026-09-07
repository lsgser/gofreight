package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MigrationKind describes how a migration name should be interpreted.
type MigrationKind struct {
	Action string // create, alter, drop
	Table  string
	Column string
}

// ParseMigrationName infers table/action from Laravel-style migration names.
func ParseMigrationName(name string) MigrationKind {
	base := strings.TrimSuffix(strings.TrimSpace(name), "_table")
	switch {
	case strings.HasPrefix(base, "create_"):
		return MigrationKind{Action: "create", Table: strings.TrimPrefix(base, "create_")}
	case strings.HasPrefix(base, "drop_"):
		return MigrationKind{Action: "drop", Table: strings.TrimPrefix(base, "drop_")}
	case strings.HasPrefix(base, "add_soft_deletes_tz_to_"):
		return MigrationKind{Action: "add_soft_deletes_tz", Table: strings.TrimPrefix(base, "add_soft_deletes_tz_to_")}
	case strings.HasPrefix(base, "add_soft_deletes_to_"):
		return MigrationKind{Action: "add_soft_deletes", Table: strings.TrimPrefix(base, "add_soft_deletes_to_")}
	case strings.HasPrefix(base, "add_timestamps_to_"):
		return MigrationKind{Action: "add_timestamps", Table: strings.TrimPrefix(base, "add_timestamps_to_")}
	case strings.HasPrefix(base, "add_remember_token_to_"):
		return MigrationKind{Action: "add_remember_token", Table: strings.TrimPrefix(base, "add_remember_token_to_")}
	case strings.HasPrefix(base, "add_") && strings.Contains(base, "_to_"):
		rest := strings.TrimPrefix(base, "add_")
		parts := strings.SplitN(rest, "_to_", 2)
		if len(parts) == 2 {
			return MigrationKind{Action: "alter", Table: parts[1], Column: parts[0]}
		}
	}
	return MigrationKind{Action: "create", Table: snakeCase(name)}
}

// BlueprintLine returns a fluent blueprint statement for a parsed field.
func (pf ParsedField) BlueprintLine() string {
	base, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(pf.RawType)), ":")
	if strings.HasSuffix(strings.ToLower(pf.RawType), ":unique") {
		base, _, _ = strings.Cut(base, ":")
	}

	var line string
	switch base {
	case "text":
		line = fmt.Sprintf(`b.Text("%s")`, pf.DBTag)
	case "int", "integer", "bigint", "references", "reference", "belongs_to":
		line = fmt.Sprintf(`b.Integer("%s")`, pf.DBTag)
	case "bool", "boolean":
		line = fmt.Sprintf(`b.Boolean("%s")`, pf.DBTag)
	case "datetime", "timestamp", "date", "time":
		line = fmt.Sprintf(`b.DateTime("%s")`, pf.DBTag)
	default:
		line = fmt.Sprintf(`b.String("%s")`, pf.DBTag)
	}

	switch base {
	case "bool", "boolean":
		line += `.Default("0")`
	case "datetime", "timestamp", "date", "time":
		// nullable by default
	default:
		line += `.NotNull()`
	}

	if pf.Unique {
		line += `.Unique()`
	}
	return line
}

// MigrationFileData holds template data for a blueprint migration file.
type MigrationFileData struct {
	Version    string
	FuncPrefix string
	Table      string
	Action     string
	Column     string
	Fields     []ParsedField
}

func migrationFuncPrefix(version, name string) string {
	s := version + "_" + name
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}, s)
	return structName(s)
}

// CreateBlueprintMigration writes a Laravel-style Go migration file.
func CreateBlueprintMigration(dir, name string, fields map[string]string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	version := time.Now().Format("20060102150405")
	base := fmt.Sprintf("%s_%s", version, name)
	kind := ParseMigrationName(name)

	var parsed []ParsedField
	for fname, ftype := range fields {
		parsed = append(parsed, ParseField(fname, ftype))
	}

	data := MigrationFileData{
		Version:    base,
		FuncPrefix: migrationFuncPrefix(base, name),
		Table:      kind.Table,
		Action:     kind.Action,
		Column:     kind.Column,
		Fields:     parsed,
	}

	path := filepath.Join(dir, base+".go")
	if err := writeTemplate(path, blueprintMigrationTmpl, data); err != nil {
		return "", err
	}
	return path, nil
}

const blueprintMigrationTmpl = `package migrate

import (
	"context"

	"github.com/lsgser/gofreight/database"
)

func init() {
	database.RegisterMigration("{{.Version}}", up{{.FuncPrefix}}, down{{.FuncPrefix}})
}

func up{{.FuncPrefix}}(ctx context.Context) error {
{{if eq .Action "create"}}	return database.SchemaCreate(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.Id()
{{range .Fields}}		{{.BlueprintLine}}
{{end}}		b.Timestamps()
	})
{{else if eq .Action "drop"}}	return database.SchemaDropIfExists(ctx, "{{.Table}}")
{{else if eq .Action "add_soft_deletes"}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.SoftDeletes()
	})
{{else if eq .Action "add_soft_deletes_tz"}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.SoftDeletesTz()
	})
{{else if eq .Action "add_timestamps"}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.Timestamps()
	})
{{else if eq .Action "add_remember_token"}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.RememberToken()
	})
{{else}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		// TODO: add columns (e.g. b.String("{{.Column}}").NotNull())
	})
{{end}}}

func down{{.FuncPrefix}}(ctx context.Context) error {
{{if eq .Action "create"}}	return database.SchemaDropIfExists(ctx, "{{.Table}}")
{{else if eq .Action "drop"}}	return database.SchemaCreate(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.Id()
		b.Timestamps()
	})
{{else if eq .Action "add_soft_deletes"}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.DropSoftDeletes()
	})
{{else if eq .Action "add_soft_deletes_tz"}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.DropSoftDeletesTz()
	})
{{else if eq .Action "add_timestamps"}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.DropTimestamps()
	})
{{else if eq .Action "add_remember_token"}}	return database.SchemaTable(ctx, "{{.Table}}", func(b *database.Blueprint) {
		b.DropRememberToken()
	})
{{else}}	// TODO: reverse alter — e.g. b.DropColumn("{{.Column}}")
	return nil
{{end}}}
`

const initialGoMigrationTmpl = `package migrate

import (
	"context"

	"github.com/lsgser/gofreight/database"
)

func init() {
	database.RegisterMigration("0001_init", up0001Init, down0001Init)
}

// up0001Init is a bootstrap migration — confirms the migration runner is wired.
func up0001Init(ctx context.Context) error {
	return nil
}

func down0001Init(ctx context.Context) error {
	return nil
}
`

const migrateToolMainTmpl = `package main

import (
	"fmt"
	"os"

	_ "{{.Module}}/db/migrate"
	"github.com/joho/godotenv"
	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/database"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if _, err := database.Connect(cfg.DatabaseURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: migrate [up|down|status|reset|refresh|fresh]")
		os.Exit(1)
	}

	m := database.NewGoMigrator()
	var err error
	switch os.Args[1] {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "status":
		var statuses []database.MigrationStatus
		statuses, err = m.Status()
		if err == nil {
			for _, s := range statuses {
				state := "Pending"
				if s.Applied {
					state = "Ran"
				}
				fmt.Printf("  [%s] %s\n", state, s.Version)
			}
		}
	case "reset":
		err = m.Reset()
	case "refresh":
		err = m.Refresh()
	case "fresh":
		err = m.Fresh()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`

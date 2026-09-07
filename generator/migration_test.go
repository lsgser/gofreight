package generator

import (
	"os"
	"strings"
	"testing"
)

func TestParseMigrationNameCreate(t *testing.T) {
	k := ParseMigrationName("create_users_table")
	if k.Action != "create" || k.Table != "users" {
		t.Fatalf("unexpected: %+v", k)
	}
}

func TestParseMigrationNameAlter(t *testing.T) {
	k := ParseMigrationName("add_email_to_users_table")
	if k.Action != "alter" || k.Table != "users" || k.Column != "email" {
		t.Fatalf("unexpected: %+v", k)
	}
}

func TestBlueprintLineUnique(t *testing.T) {
	pf := ParseField("email", "string:unique")
	line := pf.BlueprintLine()
	if !strings.Contains(line, `.Unique()`) || !strings.Contains(line, `b.String("email")`) {
		t.Fatalf("unexpected line: %s", line)
	}
}

func TestCreateBlueprintMigration(t *testing.T) {
	dir := t.TempDir()
	path, err := CreateBlueprintMigration(dir, "create_users_table", map[string]string{
		"email": "string:unique",
		"name":  "string",
	})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	for _, want := range []string{
		"package migrate",
		"database.RegisterMigration",
		"database.SchemaCreate",
		`b.String("email")`,
		`.Unique()`,
		"b.Id()",
		"b.Timestamps()",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
}

func TestParseMigrationNameSoftDeletes(t *testing.T) {
	k := ParseMigrationName("add_soft_deletes_to_posts_table")
	if k.Action != "add_soft_deletes" || k.Table != "posts" {
		t.Fatalf("unexpected: %+v", k)
	}
}

func TestCreateBlueprintMigrationSoftDeletes(t *testing.T) {
	dir := t.TempDir()
	path, err := CreateBlueprintMigration(dir, "add_soft_deletes_to_posts_table", nil)
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)
	if !strings.Contains(s, "b.SoftDeletes()") || !strings.Contains(s, "b.DropSoftDeletes()") {
		t.Fatalf("unexpected migration:\n%s", s)
	}
}

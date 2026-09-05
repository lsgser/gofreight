package generator

import (
	"strings"
	"testing"
)

func TestParseFieldEnum(t *testing.T) {
	pf := ParseField("status", "enum:draft,published,archived")
	if pf.GoType != "string" || pf.FormType != "select" {
		t.Fatalf("unexpected field: %+v", pf)
	}
	if len(pf.EnumValues) != 3 {
		t.Fatalf("enum values: %v", pf.EnumValues)
	}
	check := pf.MigrationColumnDef()
	if !strings.Contains(check, "draft") {
		t.Fatalf("check: %q", check)
	}
}

func TestParseFieldReferences(t *testing.T) {
	pf := ParseField("post_id", "references:posts")
	if pf.GoType != "int64" || pf.ReferenceTable != "posts" {
		t.Fatalf("unexpected field: %+v", pf)
	}
}

func TestParseFieldAliases(t *testing.T) {
	cases := map[string]string{
		"int":     "int",
		"integer": "int",
		"bool":    "bool",
		"jsonb":   "string",
		"decimal": "float64",
	}
	for typ, want := range cases {
		if got := ParseField("x", typ).GoType; got != want {
			t.Fatalf("%s -> %s want %s", typ, got, want)
		}
	}
}

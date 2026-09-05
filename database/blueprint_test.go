package database

import "testing"

func TestCreateTableBlueprint(t *testing.T) {
	up, down := CreateTableBlueprint("articles", func(b *Blueprint) {
		b.StringColumn("title", colNotNull())
		b.BooleanColumn("published")
	})
	if up == "" || down == "" {
		t.Fatal("expected up and down SQL")
	}
	if !contains(up, "CREATE TABLE") || !contains(up, "articles") {
		t.Fatalf("unexpected up: %s", up)
	}
	if !contains(down, "DROP TABLE") {
		t.Fatalf("unexpected down: %s", down)
	}
}

func TestAlterTableBlueprint(t *testing.T) {
	up, down := AlterTableBlueprint("posts", func(b *Blueprint) {
		b.StringColumn("slug")
	})
	if !contains(up, "ADD COLUMN") || !contains(down, "DROP COLUMN") {
		t.Fatalf("alter mismatch up=%s down=%s", up, down)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

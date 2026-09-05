package model_test

import (
	"context"
	"testing"

	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/model"
)

type SoftPost struct {
	model.Record
	Title string `db:"title" json:"title"`
}

var SoftPosts = model.NewRepository[SoftPost]("soft_posts").EnableSoftDelete()

func setupSoftDeleteTest(t *testing.T) {
	t.Helper()
	database.Reset()
	_, err := database.Connect("sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.Migrate(`CREATE TABLE soft_posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now')),
		deleted_at TEXT
	)`)
}

func TestSoftDeleteExcludesTrashed(t *testing.T) {
	setupSoftDeleteTest(t)
	ctx := context.Background()

	post := &SoftPost{Title: "Hello"}
	if err := SoftPosts.Create(ctx, post); err != nil {
		t.Fatal(err)
	}

	if err := SoftPosts.Destroy(ctx, post); err != nil {
		t.Fatal(err)
	}

	_, err := SoftPosts.Find(ctx, post.ID)
	if err == nil {
		t.Fatal("expected deleted record to be hidden from Find")
	}

	all, err := SoftPosts.Query(ctx).Get()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Fatalf("expected 0 active records, got %d", len(all))
	}
}

func TestSoftDeleteWithTrashedAndOnlyTrashed(t *testing.T) {
	setupSoftDeleteTest(t)
	ctx := context.Background()

	post := &SoftPost{Title: "Trashed"}
	SoftPosts.Create(ctx, post)
	SoftPosts.Destroy(ctx, post)

	found, err := SoftPosts.Query(ctx).WithTrashed().Find(post.ID)
	if err != nil {
		t.Fatalf("WithTrashed Find: %v", err)
	}
	if found.Title != "Trashed" {
		t.Fatalf("unexpected title %s", found.Title)
	}

	trashed, err := SoftPosts.Query(ctx).OnlyTrashed().Get()
	if err != nil {
		t.Fatal(err)
	}
	if len(trashed) != 1 {
		t.Fatalf("expected 1 trashed, got %d", len(trashed))
	}
}

func TestSoftDeleteRestore(t *testing.T) {
	setupSoftDeleteTest(t)
	ctx := context.Background()

	post := &SoftPost{Title: "Restore me"}
	SoftPosts.Create(ctx, post)
	SoftPosts.Destroy(ctx, post)

	if err := SoftPosts.Restore(ctx, post); err != nil {
		t.Fatal(err)
	}

	found, err := SoftPosts.Find(ctx, post.ID)
	if err != nil {
		t.Fatalf("after restore: %v", err)
	}
	if found.IsSoftDeleted() {
		t.Fatal("expected deleted_at to be cleared")
	}
}

func TestSoftDeleteForceDestroy(t *testing.T) {
	setupSoftDeleteTest(t)
	ctx := context.Background()

	post := &SoftPost{Title: "Gone"}
	SoftPosts.Create(ctx, post)
	SoftPosts.Destroy(ctx, post)

	if err := SoftPosts.ForceDestroy(ctx, post); err != nil {
		t.Fatal(err)
	}

	_, err := SoftPosts.Query(ctx).WithTrashed().Find(post.ID)
	if err == nil {
		t.Fatal("expected record to be permanently deleted")
	}
}

func TestSoftDeleteDeleteAll(t *testing.T) {
	setupSoftDeleteTest(t)
	ctx := context.Background()

	SoftPosts.Create(ctx, &SoftPost{Title: "A"})
	SoftPosts.Create(ctx, &SoftPost{Title: "B"})

	n, err := SoftPosts.Query(ctx).DeleteAll()
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 soft deleted, got %d", n)
	}

	count, _ := SoftPosts.Count(ctx)
	if count != 0 {
		t.Fatalf("expected 0 active, got %d", count)
	}

	trashed, _ := SoftPosts.Query(ctx).OnlyTrashed().Count()
	if trashed != 2 {
		t.Fatalf("expected 2 trashed, got %d", trashed)
	}
}

func TestSoftDeleteDisabledErrors(t *testing.T) {
	setupSoftDeleteTest(t)
	ctx := context.Background()
	repo := model.NewRepository[SoftPost]("soft_posts")

	post := &SoftPost{Title: "X"}
	repo.Create(ctx, post)

	if err := repo.Restore(ctx, post); err != model.ErrSoftDeleteDisabled {
		t.Fatalf("expected ErrSoftDeleteDisabled, got %v", err)
	}
}

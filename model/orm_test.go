package model_test

/*
|--------------------------------------------------------------------------
| Orm
|--------------------------------------------------------------------------
|
| Test suite for Orm in the model package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The model package is the ORM layer: repositories, queries, associations,
| soft deletes, validation, serialization, collections, and pagination.
| 
| Models map to tables via struct tags; migrations define schema
| separately in db/migrate.
| 
| See docs/models.md, docs/orm.md, and docs/factories.md for
| Laravel-aligned patterns.
| 
| Symbols defined here include: Article (exported type); Articles
| (exported value).
| 
| Run with go test ./model/... or go test for this package from the
| framework root.
| 
*/

import (
	"context"
	"testing"

	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/model"
)

type Article struct {
	model.Record
	Title     string `db:"title" json:"title"`
	Body      string `db:"body" json:"body"`
	Published bool   `db:"published" json:"published"`
	Views     int    `db:"views" json:"views"`
}

var Articles = model.NewRepository[Article]("articles").
	Scope("published", func(q *model.Query[Article]) *model.Query[Article] {
		return q.WhereEq("published", true)
	})

func setupORMTest(t *testing.T) {
	t.Helper()
	database.Reset()
	_, err := database.Connect("sqlite://:memory:")
	if err != nil {
		t.Fatal(err)
	}
	database.Migrate(`CREATE TABLE articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		body TEXT,
		published BOOLEAN DEFAULT 0,
		views INTEGER DEFAULT 0,
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	)`)
}

func TestQueryWhereOrderLimit(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	Articles.Create(ctx, &Article{Title: "First", Body: "Body", Published: true, Views: 10})
	Articles.Create(ctx, &Article{Title: "Second", Body: "Body", Published: false, Views: 20})
	Articles.Create(ctx, &Article{Title: "Third", Body: "Body", Published: true, Views: 30})

	results, err := Articles.Query(ctx).WhereEq("published", true).OrderDesc("views").Limit(1).Get()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1, got %d", len(results))
	}
	if results[0].Title != "Third" {
		t.Fatalf("expected Third, got %s", results[0].Title)
	}
}

func TestScope(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	Articles.Create(ctx, &Article{Title: "Pub", Body: "B", Published: true})
	Articles.Create(ctx, &Article{Title: "Draft", Body: "B", Published: false})

	results, err := Articles.Query(ctx).Scope("published").Get()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 published, got %d", len(results))
	}
}

func TestAggregates(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	Articles.Create(ctx, &Article{Title: "A", Body: "B", Views: 10})
	Articles.Create(ctx, &Article{Title: "C", Body: "B", Views: 20})

	count, err := Articles.Query(ctx).Count()
	if err != nil || count != 2 {
		t.Fatalf("count: %d, err: %v", count, err)
	}

	sum, err := Articles.Query(ctx).Sum("views")
	if err != nil || sum != 30 {
		t.Fatalf("sum: %f, err: %v", sum, err)
	}

	avg, err := Articles.Query(ctx).Avg("views")
	if err != nil || avg != 15 {
		t.Fatalf("avg: %f, err: %v", avg, err)
	}
}

func TestPluck(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	Articles.Create(ctx, &Article{Title: "Alpha", Body: "B"})
	Articles.Create(ctx, &Article{Title: "Beta", Body: "B"})

	titles, err := Articles.Query(ctx).Order("title").PluckStrings("title")
	if err != nil {
		t.Fatal(err)
	}
	if len(titles) != 2 || titles[0] != "Alpha" {
		t.Fatalf("got %v", titles)
	}
}

func TestPaginate(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		Articles.Create(ctx, &Article{Title: "Post", Body: "Body"})
	}

	page, err := Articles.Query(ctx).Paginate(2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 25 {
		t.Fatalf("total: %d", page.Total)
	}
	if page.CurrentPage != 2 {
		t.Fatalf("page: %d", page.CurrentPage)
	}
	if len(page.Data) != 10 {
		t.Fatalf("data len: %d", len(page.Data))
	}
	if page.LastPage != 3 {
		t.Fatalf("last page: %d", page.LastPage)
	}
	links := page.LinksFor("/posts")
	if links["first"] == nil {
		t.Fatalf("expected links: %+v", links)
	}
}

func TestSimplePaginate(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()
	for i := 0; i < 21; i++ {
		Articles.Create(ctx, &Article{Title: "P", Body: "b", Published: true})
	}
	page, err := Articles.Query(ctx).SimplePaginate(1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if !page.HasMorePages || len(page.Data) != 20 {
		t.Fatalf("unexpected simple page: %+v", page)
	}
}

func TestGetCollection(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()
	Articles.Create(ctx, &Article{Title: "One", Body: "b", Published: true})
	col, err := Articles.Query(ctx).GetCollection()
	if err != nil {
		t.Fatal(err)
	}
	if col.Count() != 1 {
		t.Fatalf("expected 1, got %d", col.Count())
	}
}

func TestFindOrCreate(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	record, created, err := Articles.Query(ctx).FirstOrCreate(map[string]any{
		"title": "New Post",
		"body":  "Content here",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("expected created=true")
	}

	_, created2, err := Articles.Query(ctx).FirstOrCreate(map[string]any{
		"title": "New Post",
		"body":  "Content here",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created2 {
		t.Fatal("expected created=false on second call")
	}
	if record.ID == 0 {
		t.Fatal("expected ID assigned")
	}
}

func TestUpdateAllDeleteAll(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	Articles.Create(ctx, &Article{Title: "A", Body: "B", Published: false})
	Articles.Create(ctx, &Article{Title: "C", Body: "B", Published: false})

	affected, err := Articles.Query(ctx).WhereEq("published", false).UpdateAll(map[string]any{"published": true})
	if err != nil || affected != 2 {
		t.Fatalf("update all: %d, %v", affected, err)
	}

	affected, err = Articles.Query(ctx).WhereEq("published", true).DeleteAll()
	if err != nil || affected != 2 {
		t.Fatalf("delete all: %d, %v", affected, err)
	}
}

func TestTransaction(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	err := model.Transaction(ctx, func(txCtx context.Context) error {
		return Articles.Create(txCtx, &Article{Title: "In TX", Body: "Body"})
	})
	if err != nil {
		t.Fatal(err)
	}

	count, _ := Articles.Count(ctx)
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}
}

func TestDirtyTracking(t *testing.T) {
	post := &Article{Title: "Original", Body: "Body"}
	d := &model.Dirty{}
	d.Snapshot(post)

	post.Title = "Changed"
	if !d.Changed() {
		t.Fatal("expected changed")
	}
	fields := d.ChangedFields()
	if len(fields) != 1 || fields[0] != "title" {
		t.Fatalf("changed fields: %v", fields)
	}
}

func TestWhereIn(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	Articles.Create(ctx, &Article{Title: "A", Body: "B"})
	Articles.Create(ctx, &Article{Title: "B", Body: "B"})
	Articles.Create(ctx, &Article{Title: "C", Body: "B"})

	results, err := Articles.Query(ctx).WhereIn("title", []any{"A", "C"}).Get()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestFindEach(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		Articles.Create(ctx, &Article{Title: "Post", Body: "B"})
	}

	var count int
	err := Articles.Query(ctx).FindEach(2, func(a Article) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Fatalf("expected 5, got %d", count)
	}
}

func TestToSQL(t *testing.T) {
	setupORMTest(t)
	ctx := context.Background()

	sql, args := Articles.Query(ctx).WhereEq("published", true).OrderDesc("id").Limit(5).ToSQL()
	if sql == "" {
		t.Fatal("expected SQL")
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}

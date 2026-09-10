package model_test

/*
|--------------------------------------------------------------------------
| Collection
|--------------------------------------------------------------------------
|
| Test suite for Collection in the model package.
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
| Run with go test ./model/... or go test for this package from the
| framework root.
| 
*/

import (
	"context"
	"testing"

	"github.com/lsgser/gofreight/model"
)

type seedA struct{ model.Seeder }

func (s *seedA) Run(ctx context.Context) error { return nil }

type seedB struct {
	model.Seeder
	called *bool
}

func (s *seedB) Run(ctx context.Context) error {
	*s.called = true
	return nil
}

func TestSeederCall(t *testing.T) {
	var bCalled bool
	base := model.NewSeeder()
	base.SetContext(context.Background())
	if err := base.Call(&seedA{}, &seedB{called: &bCalled}); err != nil {
		t.Fatal(err)
	}
	if !bCalled {
		t.Fatal("expected seedB to run")
	}
}

func TestCollectionFindPluck(t *testing.T) {
	items := []Article{
		{Record: model.Record{ID: 1}, Title: "A"},
		{Record: model.Record{ID: 2}, Title: "B"},
	}
	col := model.NewCollection(items)

	if _, ok := col.Find(2); !ok {
		t.Fatal("expected find 2")
	}
	keys := col.ModelKeys()
	if len(keys) != 2 || keys[0] != 1 {
		t.Fatalf("keys: %v", keys)
	}
	plucked := col.Pluck("title")
	if len(plucked) != 2 || plucked[0] != "A" {
		t.Fatalf("pluck: %v", plucked)
	}
	filtered := col.Filter(func(a Article) bool { return a.Title == "B" })
	if filtered.Count() != 1 {
		t.Fatal("filter failed")
	}
}

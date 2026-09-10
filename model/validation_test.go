package model_test

/*
|--------------------------------------------------------------------------
| Validation
|--------------------------------------------------------------------------
|
| Test suite for Validation in the model package.
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
| Symbols defined here include: TestRecord (exported type).
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

type TestRecord struct {
	model.Record
	Title string
	Body  string
}

func (r *TestRecord) Validators() []model.Validator {
	return []model.Validator{
		model.Presence("Title"),
		model.Length("Title", 3, 100),
		model.Presence("Body"),
	}
}

func TestValidatePresence(t *testing.T) {
	record := &TestRecord{}
	errs := model.Validate(record)
	if !errs.Any() {
		t.Fatal("expected validation errors")
	}
	if _, ok := errs["title"]; !ok {
		t.Fatal("expected title error")
	}
}

func TestValidateLength(t *testing.T) {
	record := &TestRecord{Title: "ab", Body: "valid body here"}
	errs := model.Validate(record)
	if !errs.Any() {
		t.Fatal("expected length error for title")
	}
}

func TestValidatePass(t *testing.T) {
	record := &TestRecord{Title: "Valid Title", Body: "Valid body content"}
	errs := model.Validate(record)
	if errs.Any() {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestCallbacks(t *testing.T) {
	cbs := model.NewCallbacks()
	var called bool
	cbs.Register(model.BeforeSave, func(ctx context.Context, record any) error {
		called = true
		return nil
	})

	record := &TestRecord{}
	if err := cbs.Run(context.Background(), model.BeforeSave, record); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected callback to be called")
	}
}

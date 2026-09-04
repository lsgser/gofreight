package model_test

import (
	"context"
	"testing"

	"github.com/gofreight/gofreight/model"
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

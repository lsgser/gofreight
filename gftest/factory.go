package gftest

import (
	"context"
	"testing"

	"github.com/lsgser/gofreight/model"
)

// Factory creates test model instances, like Laravel factories / FactoryBot.
type Factory[T any] struct {
	repo     *model.Repository[T]
	defaults map[string]any
	sequence map[string]int
}

// NewFactory creates a model factory for the given repository.
func NewFactory[T any](repo *model.Repository[T]) *Factory[T] {
	return &Factory[T]{
		repo:     repo,
		defaults: make(map[string]any),
		sequence: make(map[string]int),
	}
}

// Define sets default attribute values for the factory.
func (f *Factory[T]) Define(defaults map[string]any) *Factory[T] {
	for k, v := range defaults {
		f.defaults[k] = v
	}
	return f
}

// Sequence sets a sequenced value for a field (increments each call).
func (f *Factory[T]) Sequence(field string, fn func(n int) any) *Factory[T] {
	n := f.sequence[field]
	f.defaults[field] = fn(n)
	f.sequence[field] = n + 1
	return f
}

// Make builds an unsaved model instance.
func (f *Factory[T]) Make(overrides ...map[string]any) *T {
	attrs := f.mergeAttrs(overrides...)
	var record T
	setAttrs(&record, attrs)
	return &record
}

// Create builds and persists a model instance.
func (f *Factory[T]) Create(t *testing.T, overrides ...map[string]any) *T {
	t.Helper()
	record := f.Make(overrides...)
	if err := f.repo.Save(context.Background(), record); err != nil {
		t.Fatalf("factory create: %v", err)
	}
	return record
}

// CreateMany creates n persisted records.
func (f *Factory[T]) CreateMany(t *testing.T, n int, overrides ...map[string]any) []*T {
	t.Helper()
	records := make([]*T, n)
	for i := 0; i < n; i++ {
		records[i] = f.Create(t, overrides...)
	}
	return records
}

// RawCreate saves a record without validation (for edge case tests).
func (f *Factory[T]) RawCreate(t *testing.T, overrides ...map[string]any) *T {
	t.Helper()
	record := f.Make(overrides...)
	if err := f.repo.Create(context.Background(), record); err != nil {
		t.Fatalf("factory raw create: %v", err)
	}
	return record
}

func (f *Factory[T]) mergeAttrs(overrides ...map[string]any) map[string]any {
	attrs := make(map[string]any)
	for k, v := range f.defaults {
		attrs[k] = v
	}
	for _, o := range overrides {
		for k, v := range o {
			attrs[k] = v
		}
	}
	return attrs
}

package gftest

import (
	"context"
	"maps"
	"testing"

	"github.com/lsgser/gofreight/model"
)

type sequenceCounter struct {
	counts map[string]int
}

// Factory creates test model instances with default attributes.
type Factory[T any] struct {
	repo          *model.Repository[T]
	defaults      map[string]any
	sequenceFns   map[string]func(n int) any
	seq           *sequenceCounter
	states        []map[string]any
	stateFns      []func(map[string]any) map[string]any
	afterMaking   []func(*T)
	afterCreating []func(*T)
}

// NewFactory creates a model factory for the given repository.
func NewFactory[T any](repo *model.Repository[T]) *Factory[T] {
	return &Factory[T]{
		repo:     repo,
		defaults: make(map[string]any),
		seq:      &sequenceCounter{counts: make(map[string]int)},
	}
}

// Define sets default attribute values for the factory.
func (f *Factory[T]) Define(defaults map[string]any) *Factory[T] {
	c := f.clone()
	for k, v := range defaults {
		c.defaults[k] = v
	}
	return c
}

// Sequence sets a sequenced value for a field (increments each Make/Create).
func (f *Factory[T]) Sequence(field string, fn func(n int) any) *Factory[T] {
	c := f.clone()
	if c.sequenceFns == nil {
		c.sequenceFns = make(map[string]func(n int) any)
	}
	c.sequenceFns[field] = fn
	return c
}

// State applies attribute overrides (Laravel factory states).
func (f *Factory[T]) State(attrs map[string]any) *Factory[T] {
	c := f.clone()
	c.states = append(c.states, attrs)
	return c
}

// StateFn applies a state callback (Laravel state(fn)).
func (f *Factory[T]) StateFn(fn func(attrs map[string]any) map[string]any) *Factory[T] {
	c := f.clone()
	c.stateFns = append(c.stateFns, fn)
	return c
}

// AfterMaking registers a callback after building an instance.
func (f *Factory[T]) AfterMaking(fn func(*T)) *Factory[T] {
	c := f.clone()
	c.afterMaking = append(c.afterMaking, fn)
	return c
}

// AfterCreating registers a callback after persisting an instance.
func (f *Factory[T]) AfterCreating(fn func(*T)) *Factory[T] {
	c := f.clone()
	c.afterCreating = append(c.afterCreating, fn)
	return c
}

// Count returns a counter for fluent bulk creation (Laravel count(n)).
func (f *Factory[T]) Count(n int) *FactoryCounter[T] {
	return &FactoryCounter[T]{f: f, n: n}
}

// FactoryCounter creates multiple records via Make or Create.
type FactoryCounter[T any] struct {
	f *Factory[T]
	n int
}

// Make builds n unsaved instances.
func (c *FactoryCounter[T]) Make(overrides ...map[string]any) []*T {
	out := make([]*T, c.n)
	for i := range out {
		out[i] = c.f.Make(overrides...)
	}
	return out
}

// Create builds and persists n instances.
func (c *FactoryCounter[T]) Create(t *testing.T, overrides ...map[string]any) []*T {
	t.Helper()
	out := make([]*T, c.n)
	for i := range out {
		out[i] = c.f.Create(t, overrides...)
	}
	return out
}

// Make builds an unsaved model instance.
func (f *Factory[T]) Make(overrides ...map[string]any) *T {
	attrs := f.mergeAttrs(overrides...)
	var record T
	setAttrs(&record, attrs)
	for _, fn := range f.afterMaking {
		fn(&record)
	}
	return &record
}

// Create builds and persists a model instance.
func (f *Factory[T]) Create(t *testing.T, overrides ...map[string]any) *T {
	t.Helper()
	record := f.Make(overrides...)
	if err := f.repo.Save(context.Background(), record); err != nil {
		t.Fatalf("factory create: %v", err)
	}
	for _, fn := range f.afterCreating {
		fn(record)
	}
	return record
}

// CreateMany creates n persisted records.
func (f *Factory[T]) CreateMany(t *testing.T, n int, overrides ...map[string]any) []*T {
	return f.Count(n).Create(t, overrides...)
}

// RawCreate saves a record without validation (for edge case tests).
func (f *Factory[T]) RawCreate(t *testing.T, overrides ...map[string]any) *T {
	t.Helper()
	record := f.Make(overrides...)
	if err := f.repo.Create(context.Background(), record); err != nil {
		t.Fatalf("factory raw create: %v", err)
	}
	for _, fn := range f.afterCreating {
		fn(record)
	}
	return record
}

func (f *Factory[T]) clone() *Factory[T] {
	c := &Factory[T]{
		repo:          f.repo,
		defaults:      maps.Clone(f.defaults),
		sequenceFns:   maps.Clone(f.sequenceFns),
		seq:           f.seq,
		states:        append([]map[string]any{}, f.states...),
		stateFns:      append([]func(map[string]any) map[string]any{}, f.stateFns...),
		afterMaking:   append([]func(*T){}, f.afterMaking...),
		afterCreating: append([]func(*T){}, f.afterCreating...),
	}
	return c
}

func (f *Factory[T]) mergeAttrs(overrides ...map[string]any) map[string]any {
	attrs := make(map[string]any)
	for k, v := range f.defaults {
		attrs[k] = resolveAttr(v)
	}
	for field, fn := range f.sequenceFns {
		n := f.seq.counts[field]
		attrs[field] = fn(n)
		f.seq.counts[field] = n + 1
	}
	for _, state := range f.states {
		for k, v := range state {
			attrs[k] = resolveAttr(v)
		}
	}
	for _, fn := range f.stateFns {
		merged := fn(maps.Clone(attrs))
		for k, v := range merged {
			attrs[k] = resolveAttr(v)
		}
	}
	for _, o := range overrides {
		for k, v := range o {
			attrs[k] = resolveAttr(v)
		}
	}
	return attrs
}

func resolveAttr(v any) any {
	if fn, ok := v.(func() any); ok {
		return fn()
	}
	return v
}

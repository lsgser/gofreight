package model

import (
	"encoding/json"
	"reflect"
)

// Collection wraps model query results (Laravel Eloquent Collection).
type Collection[T any] struct {
	items []T
}

// NewCollection creates a collection from a slice.
func NewCollection[T any](items []T) Collection[T] {
	if items == nil {
		items = []T{}
	}
	return Collection[T]{items: items}
}

// All returns the underlying slice.
func (c Collection[T]) All() []T {
	return c.items
}

// Count returns the number of models in the collection.
func (c Collection[T]) Count() int {
	return len(c.items)
}

// IsEmpty reports whether the collection has no items.
func (c Collection[T]) IsEmpty() bool {
	return len(c.items) == 0
}

// First returns the first model, if any.
func (c Collection[T]) First() (*T, bool) {
	if len(c.items) == 0 {
		return nil, false
	}
	return &c.items[0], true
}

// Last returns the last model, if any.
func (c Collection[T]) Last() (*T, bool) {
	if len(c.items) == 0 {
		return nil, false
	}
	return &c.items[len(c.items)-1], true
}

// Find returns the model with the given primary key.
func (c Collection[T]) Find(id int64) (*T, bool) {
	for i := range c.items {
		if recordID(&c.items[i]) == id {
			return &c.items[i], true
		}
	}
	return nil, false
}

// FindOrFail returns the model with the given primary key or panics.
func (c Collection[T]) FindOrFail(id int64) *T {
	item, ok := c.Find(id)
	if !ok {
		panic("model: record not found in collection")
	}
	return item
}

// Contains reports whether a model with the given primary key exists.
func (c Collection[T]) Contains(id int64) bool {
	_, ok := c.Find(id)
	return ok
}

// ModelKeys returns primary keys for all models in the collection.
func (c Collection[T]) ModelKeys() []int64 {
	keys := make([]int64, len(c.items))
	for i := range c.items {
		keys[i] = recordID(&c.items[i])
	}
	return keys
}

// Filter returns models matching the predicate.
func (c Collection[T]) Filter(fn func(T) bool) Collection[T] {
	out := make([]T, 0, len(c.items))
	for _, item := range c.items {
		if fn(item) {
			out = append(out, item)
		}
	}
	return NewCollection(out)
}

// Map transforms each model in the collection.
func (c Collection[T]) Map(fn func(T) T) Collection[T] {
	out := make([]T, len(c.items))
	for i, item := range c.items {
		out[i] = fn(item)
	}
	return NewCollection(out)
}

// Each iterates the collection.
func (c Collection[T]) Each(fn func(T) error) error {
	for _, item := range c.items {
		if err := fn(item); err != nil {
			return err
		}
	}
	return nil
}

// Pluck returns values for a struct field by db tag or exported name.
func (c Collection[T]) Pluck(field string) []any {
	if len(c.items) == 0 {
		return nil
	}
	out := make([]any, len(c.items))
	for i := range c.items {
		out[i] = fieldValue(&c.items[i], field)
	}
	return out
}

// Only returns models whose primary keys are in the given set.
func (c Collection[T]) Only(ids ...int64) Collection[T] {
	set := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return c.Filter(func(item T) bool {
		_, ok := set[recordID(&item)]
		return ok
	})
}

// Except returns models whose primary keys are not in the given set.
func (c Collection[T]) Except(ids ...int64) Collection[T] {
	set := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return c.Filter(func(item T) bool {
		_, ok := set[recordID(&item)]
		return !ok
	})
}

// ToArray converts the collection to a slice of maps (Laravel Collection::toArray).
func (c Collection[T]) ToArray() []map[string]any {
	out := make([]map[string]any, len(c.items))
	for i := range c.items {
		out[i] = ToArray(&c.items[i])
	}
	return out
}

// ToJSON converts the collection to JSON (Laravel Collection JSON encoding).
func (c Collection[T]) ToJSON() ([]byte, error) {
	return json.Marshal(c.ToArray())
}

func fieldValue(record any, field string) any {
	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("db")
		if tag == field || f.Name == field {
			return v.Field(i).Interface()
		}
	}
	return nil
}

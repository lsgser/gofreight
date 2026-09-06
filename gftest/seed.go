package gftest

import (
	"context"
	"testing"
)

// Seeder defines a database seeder run via the CLI or tests.
type Seeder interface {
	Run(ctx context.Context) error
}

// Seed runs one or more seeders.
func Seed(t *testing.T, seeders ...Seeder) {
	t.Helper()
	ctx := context.Background()
	for _, s := range seeders {
		if err := s.Run(ctx); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

// SeederFunc adapts a function to the Seeder interface.
type SeederFunc func(ctx context.Context) error

func (f SeederFunc) Run(ctx context.Context) error {
	return f(ctx)
}

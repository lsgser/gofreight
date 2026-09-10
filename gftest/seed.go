package gftest

/*
|--------------------------------------------------------------------------
| Seed
|--------------------------------------------------------------------------
|
| Implements Seed as part of the gftest package in the Gofreight
| framework. Key symbols: Seeder, Seed, SeederFunc, Run.
| 
| gftest is the feature testing harness used from tests/ in your
| application.
| 
| NewApp boots a test HTTP server, runs migrations, exposes HTTP helpers
| (Get, Post, AssertOk), database assertions, and fakes for mail, cache,
| and queue.
| 
| Run the suite with gofreight test; see docs/testing.md for factories and
| authentication in tests.
| 
| Symbols defined here include: Seeder (exported type); Seed (Seed runs
| one or more seeders.); SeederFunc (exported type).
| 
*/

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

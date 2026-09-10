package gftest_test

/*
|--------------------------------------------------------------------------
| Gftest
|--------------------------------------------------------------------------
|
| Test suite for Gftest in the gftest package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
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
| Run with go test ./gftest/... or go test for this package from the
| framework root.
| 
*/

import (
	"testing"

	"github.com/lsgser/gofreight/gftest"
)

func TestDescribeIt(t *testing.T) {
	var ran bool
	gftest.Describe(t, "Example", func(d *gftest.DescribeContext) {
		d.It("runs the test", func(t *testing.T) {
			ran = true
		})
	})
	if !ran {
		t.Fatal("test did not run")
	}
}

func TestBeforeEach(t *testing.T) {
	count := 0
	gftest.Describe(t, "Hooks", func(d *gftest.DescribeContext) {
		d.BeforeEach(func(t *testing.T) { count++ })
		d.It("first", func(t *testing.T) {})
		d.It("second", func(t *testing.T) {})
	})
	if count != 2 {
		t.Fatalf("expected beforeEach 2 times, got %d", count)
	}
}

func TestExpectEqual(t *testing.T) {
	gftest.Expect(42).Bind(t).ToEqual(42)
	gftest.Expect("hello").Bind(t).ToContain("ell")
}

func TestExpectNegated(t *testing.T) {
	gftest.Expect(1).Bind(t).Not().ToEqual(2)
}

func TestDatasets(t *testing.T) {
	ran := 0
	gftest.Describe(t, "Datasets", func(d *gftest.DescribeContext) {
		d.ItWith("works", gftest.Datasets(map[string]map[string]any{
			"a": {"x": 1},
			"b": {"x": 2},
		}), func(t *testing.T, ds gftest.Dataset) {
			ran++
		})
	})
	if ran != 2 {
		t.Fatalf("expected 2 dataset runs, got %d", ran)
	}
}

func TestFactoryMake(t *testing.T) {
	type Article struct {
		Title string `db:"title"`
		Body  string `db:"body"`
	}

	f := gftest.NewFactory[Article](nil).Define(map[string]any{
		"title": func() any { return "Hello" },
		"body":  func() any { return "World" },
	})

	article := f.Make()
	if article.Title != "Hello" {
		t.Fatalf("got title %q", article.Title)
	}
}

func TestFactoryLazyAttrs(t *testing.T) {
	type Article struct {
		Title string `db:"title"`
	}

	f := gftest.NewFactory[Article](nil).Define(map[string]any{
		"title": func() any { return "A" },
	})
	a := f.Make(map[string]any{"title": func() any { return "B" }})
	if a.Title != "B" {
		t.Fatalf("got title %q", a.Title)
	}
}

func TestFactoryState(t *testing.T) {
	type User struct {
		Email    string `db:"email"`
		Verified bool   `db:"verified"`
	}

	unverified := gftest.NewFactory[User](nil).Define(map[string]any{
		"email":    "a@example.com",
		"verified": true,
	}).State(map[string]any{"verified": false})

	u := unverified.Make()
	if u.Verified {
		t.Fatal("state should set verified false")
	}
}

func TestFactoryStateFn(t *testing.T) {
	type User struct {
		Email string `db:"email"`
	}

	f := gftest.NewFactory[User](nil).Define(map[string]any{
		"email": "base@example.com",
	}).StateFn(func(attrs map[string]any) map[string]any {
		attrs["email"] = "unverified@example.com"
		return attrs
	})

	u := f.Make()
	if u.Email != "unverified@example.com" {
		t.Fatalf("got %q", u.Email)
	}
}

func TestFactorySequence(t *testing.T) {
	type Post struct {
		Title string `db:"title"`
	}

	f := gftest.NewFactory[Post](nil).Sequence("title", func(n int) any {
		return "Post " + string(rune('A'+n))
	})

	a := f.Make()
	b := f.Make()
	if a.Title != "Post A" || b.Title != "Post B" {
		t.Fatalf("sequence got %q and %q", a.Title, b.Title)
	}
}

func TestFactoryCount(t *testing.T) {
	type Item struct {
		Name string `db:"name"`
	}

	f := gftest.NewFactory[Item](nil).Define(map[string]any{"name": "x"})
	items := f.Count(3).Make()
	if len(items) != 3 {
		t.Fatalf("expected 3, got %d", len(items))
	}
}

func TestFactoryAfterMaking(t *testing.T) {
	type Item struct {
		Name string `db:"name"`
	}

	called := false
	f := gftest.NewFactory[Item](nil).Define(map[string]any{"name": "x"}).AfterMaking(func(i *Item) {
		called = true
		i.Name = "y"
	})

	item := f.Make()
	if !called || item.Name != "y" {
		t.Fatalf("afterMaking: called=%v name=%q", called, item.Name)
	}
}

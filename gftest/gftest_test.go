package gftest_test

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
		"title": "Hello",
		"body":  "World",
	})

	article := f.Make()
	if article.Title != "Hello" {
		t.Fatalf("got title %q", article.Title)
	}
}

package gftest

/*
|--------------------------------------------------------------------------
| Suite
|--------------------------------------------------------------------------
|
| Implements Suite as part of the gftest package in the Gofreight
| framework. Key symbols: DescribeContext, Describe, BeforeAll, AfterAll,
| BeforeEach, AfterEach.
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
| Symbols defined here include: DescribeContext (exported type); Describe
| (Describe groups related tests with shared setup hooks.); BeforeAll
| (BeforeAll runs once before all tests in the group.); AfterAll (AfterAll
| runs once after all tests in the group.); BeforeEach (BeforeEach runs
| before each test in the group.); AfterEach (AfterEach runs after each
| test in the group.); It (It registers a test case within the describe
| block.); ItWith (ItWith runs a test for each dataset entry (like Pest
| datasets).).
| 
*/

import (
	"testing"
)

// DescribeContext holds hooks and runs grouped tests (like Pest/RSpec describe blocks).
type DescribeContext struct {
	t          *testing.T
	name       string
	beforeAll  []func(*testing.T)
	afterAll   []func(*testing.T)
	beforeEach []func(*testing.T)
	afterEach  []func(*testing.T)
	tests      []namedTest
}

type namedTest struct {
	name string
	fn   func(*testing.T)
}

// Describe groups related tests with shared setup hooks.
//
//	gftest.Describe(t, "Posts", func(d *gftest.DescribeContext) {
//	    d.BeforeEach(func(t *testing.T) { ... })
//	    d.It("lists posts", func(t *testing.T) { ... })
//	})
func Describe(t *testing.T, name string, fn func(*DescribeContext)) {
	t.Helper()
	d := &DescribeContext{t: t, name: name}
	fn(d)
	d.run()
}

// BeforeAll runs once before all tests in the group.
func (d *DescribeContext) BeforeAll(fn func(*testing.T)) {
	d.beforeAll = append(d.beforeAll, fn)
}

// AfterAll runs once after all tests in the group.
func (d *DescribeContext) AfterAll(fn func(*testing.T)) {
	d.afterAll = append(d.afterAll, fn)
}

// BeforeEach runs before each test in the group.
func (d *DescribeContext) BeforeEach(fn func(*testing.T)) {
	d.beforeEach = append(d.beforeEach, fn)
}

// AfterEach runs after each test in the group.
func (d *DescribeContext) AfterEach(fn func(*testing.T)) {
	d.afterEach = append(d.afterEach, fn)
}

// It registers a test case within the describe block.
func (d *DescribeContext) It(name string, fn func(*testing.T)) {
	d.tests = append(d.tests, namedTest{name: name, fn: fn})
}

// ItWith runs a test for each dataset entry (like Pest datasets).
func (d *DescribeContext) ItWith(name string, datasets []Dataset, fn func(*testing.T, Dataset)) {
	for _, ds := range datasets {
		ds := ds
		testName := name + " [" + ds.Name + "]"
		d.tests = append(d.tests, namedTest{
			name: testName,
			fn: func(t *testing.T) {
				fn(t, ds)
			},
		})
	}
}

func (d *DescribeContext) run() {
	if len(d.tests) == 0 {
		return
	}

	// BeforeAll runs in the parent test context
	for _, fn := range d.beforeAll {
		fn(d.t)
	}

	for _, tc := range d.tests {
		d.t.Run(tc.name, func(t *testing.T) {
			for _, fn := range d.beforeEach {
				fn(t)
			}

			tc.fn(t)

			for i := len(d.afterEach) - 1; i >= 0; i-- {
				d.afterEach[i](t)
			}
		})
	}

	for _, fn := range d.afterAll {
		fn(d.t)
	}
}

// Dataset represents a single parameterized test case.
type Dataset struct {
	Name string
	Data map[string]any
}

// Datasets builds datasets from a map of name → data.
func Datasets(items map[string]map[string]any) []Dataset {
	var result []Dataset
	for name, data := range items {
		result = append(result, Dataset{Name: name, Data: data})
	}
	return result
}

// Test is a shorthand for a single It block without Describe.
func Test(t *testing.T, name string, fn func(*testing.T)) {
	t.Helper()
	t.Run(name, fn)
}

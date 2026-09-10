package gftest

/*
|--------------------------------------------------------------------------
| Expect
|--------------------------------------------------------------------------
|
| Implements Expect as part of the gftest package in the Gofreight
| framework. Key symbols: Expectation, Expect, Not, ToEqual, ToBeTrue,
| ToBeFalse.
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
| Symbols defined here include: Expectation (exported type); Expect
| (Expect creates a fluent assertion on a value.); Not (Not negates the
| next assertion.); ToEqual (ToEqual asserts deep equality.); ToBeTrue
| (ToBeTrue asserts the value is true.); ToBeFalse (ToBeFalse asserts the
| value is false.); ToBeNil (ToBeNil asserts the value is nil.); ToBeEmpty
| (ToBeEmpty asserts strings, slices, maps, or channels are empty.).
| 
*/

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Expectation provides fluent assertions inspired by Pest/Jest.
type Expectation struct {
	t        *testing.T
	actual   any
	negated  bool
	helper   bool
}

// Expect creates a fluent assertion on a value.
//
//	gftest.Expect(response.Code).ToEqual(200)
//	gftest.Expect(body).ToContain("Hello")
func Expect(actual any) *Expectation {
	return &Expectation{actual: actual, helper: true}
}

// Not negates the next assertion.
func (e *Expectation) Not() *Expectation {
	e.negated = !e.negated
	return e
}

func (e *Expectation) fail(format string, args ...any) {
	if e.t == nil {
		panic(fmt.Sprintf(format, args...))
	}
	e.t.Helper()
	e.t.Fatalf(format, args...)
}

func (e *Expectation) check(pass bool, msg string, args ...any) {
	if pass != e.negated {
		return
	}
	if e.negated {
		e.fail("expected assertion to fail: "+msg, args...)
	} else {
		e.fail(msg, args...)
	}
}

// ToEqual asserts deep equality.
func (e *Expectation) ToEqual(expected any) *Expectation {
	pass := reflect.DeepEqual(e.actual, expected)
	e.check(pass, "expected %v to equal %v", e.actual, expected)
	return e
}

// ToBeTrue asserts the value is true.
func (e *Expectation) ToBeTrue() *Expectation {
	pass, ok := e.actual.(bool)
	e.check(ok && pass, "expected %v to be true", e.actual)
	return e
}

// ToBeFalse asserts the value is false.
func (e *Expectation) ToBeFalse() *Expectation {
	pass, ok := e.actual.(bool)
	e.check(ok && !pass, "expected %v to be false", e.actual)
	return e
}

// ToBeNil asserts the value is nil.
func (e *Expectation) ToBeNil() *Expectation {
	e.check(e.actual == nil, "expected nil, got %v", e.actual)
	return e
}

// ToBeEmpty asserts strings, slices, maps, or channels are empty.
func (e *Expectation) ToBeEmpty() *Expectation {
	pass := isEmpty(e.actual)
	e.check(pass, "expected %v to be empty", e.actual)
	return e
}

// ToContain asserts a string contains a substring or slice contains an element.
func (e *Expectation) ToContain(substr any) *Expectation {
	pass := contains(e.actual, substr)
	e.check(pass, "expected %v to contain %v", e.actual, substr)
	return e
}

// ToHaveCount asserts a slice/map/string length.
func (e *Expectation) ToHaveCount(n int) *Expectation {
	pass := countOf(e.actual) == n
	e.check(pass, "expected count %d, got %d for %v", n, countOf(e.actual), e.actual)
	return e
}

// ToBeGreaterThan asserts numeric comparison.
func (e *Expectation) ToBeGreaterThan(n float64) *Expectation {
	pass := toFloat(e.actual) > n
	e.check(pass, "expected %v to be > %g", e.actual, n)
	return e
}

// ToBeLessThan asserts numeric comparison.
func (e *Expectation) ToBeLessThan(n float64) *Expectation {
	pass := toFloat(e.actual) < n
	e.check(pass, "expected %v to be < %g", e.actual, n)
	return e
}

// ToMatch asserts a string matches a substring (simple contains for regex-free API).
func (e *Expectation) ToMatch(pattern string) *Expectation {
	s, ok := e.actual.(string)
	e.check(ok && strings.Contains(s, pattern), "expected %q to match %q", s, pattern)
	return e
}

// ToBeType asserts the type name matches.
func (e *Expectation) ToBeType(typeName string) *Expectation {
	if e.actual == nil {
		e.check(false, "expected type %s, got nil", typeName)
		return e
	}
	got := reflect.TypeOf(e.actual).String()
	e.check(got == typeName || strings.HasSuffix(got, typeName), "expected type %s, got %s", typeName, got)
	return e
}

// ToPanic asserts a function panics (use with care).
func (e *Expectation) ToPanic() *Expectation {
	fn, ok := e.actual.(func())
	if !ok {
		e.fail("ToPanic requires func()")
		return e
	}
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		fn()
	}()
	e.check(panicked, "expected function to panic")
	return e
}

// ToBeError asserts the value is a non-nil error.
func (e *Expectation) ToBeError() *Expectation {
	err, ok := e.actual.(error)
	e.check(ok && err != nil, "expected error, got %v", e.actual)
	return e
}

// ToBeNoError asserts the value is nil error.
func (e *Expectation) ToBeNoError() *Expectation {
	if err, ok := e.actual.(error); ok {
		e.check(err == nil, "expected no error, got %v", err)
	} else {
		e.check(true, "")
	}
	return e
}

// Bind attaches a *testing.T for failure reporting.
func (e *Expectation) Bind(t *testing.T) *Expectation {
	e.t = t
	return e
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		return rv.Len() == 0
	default:
		return false
	}
}

func contains(actual, substr any) bool {
	switch a := actual.(type) {
	case string:
		s, ok := substr.(string)
		return ok && strings.Contains(a, s)
	default:
		rv := reflect.ValueOf(actual)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return false
		}
		for i := 0; i < rv.Len(); i++ {
			if reflect.DeepEqual(rv.Index(i).Interface(), substr) {
				return true
			}
		}
		return false
	}
}

func countOf(v any) int {
	if v == nil {
		return 0
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		return rv.Len()
	default:
		return 0
	}
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	case float32:
		return float64(n)
	default:
		return 0
	}
}

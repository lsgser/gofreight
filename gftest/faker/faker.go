// Package faker provides fake data helpers for tests and seeders.
package faker

/*
|--------------------------------------------------------------------------
| Faker
|--------------------------------------------------------------------------
|
| Implements Faker as part of the faker package in the Gofreight
| framework. Key symbols: Name, FirstName, LastName, Email, Username,
| Password.
| 
| Test fakers generate random emails, names, and type-aware values for
| factories and gftest seed data.
| 
| Used by gofreight make:factory and table-driven tests that need
| realistic but non-production data.
| 
| Symbols defined here include: Lazy (Lazy wraps a generator so factories
| produce fresh values on each Create.); DefinitionsForModel
| (DefinitionsForModel returns lazy default attributes for common model
| names.); ForFieldType (ForFieldType returns a lazy generator for a
| scaffold field type.).
| 
*/

import (
	"strings"

	gfaker "github.com/go-faker/faker/v4"

	"github.com/lsgser/gofreight/support/datetime"
)

func Name() string        { return gfaker.Name() }
func FirstName() string   { return gfaker.FirstName() }
func LastName() string    { return gfaker.LastName() }
func Email() string       { return gfaker.Email() }
func Username() string    { return gfaker.Username() }
func Password() string    { return gfaker.Password() }
func Phone() string       { return gfaker.Phonenumber() }
func URL() string         { return gfaker.URL() }
func Word() string        { return gfaker.Word() }
func Sentence() string    { return gfaker.Sentence() }
func Paragraph() string   { return gfaker.Paragraph() }
func UUID() string        { return gfaker.UUIDHyphenated() }
func Date() string        { return gfaker.Date() }
func Time() string        { return gfaker.TimeString() }
func DateTime() string    { return datetime.Now().ToDateTimeString() }

func Int() int {
	n, err := gfaker.RandomInt(1, 10_000)
	if err != nil || len(n) == 0 {
		return 1
	}
	return n[0]
}

func Int64() int64 { return int64(Int()) }

func Float() float64 {
	n, err := gfaker.RandomInt(1, 1000)
	if err != nil || len(n) == 0 {
		return 1.0
	}
	return float64(n[0]) / 10
}

func Bool() bool {
	n, err := gfaker.RandomInt(0, 1)
	if err != nil || len(n) == 0 {
		return true
	}
	return n[0] == 1
}

// Lazy wraps a generator so factories produce fresh values on each Create.
func Lazy(fn func() any) func() any { return fn }

// DefinitionsForModel returns lazy default attributes for common model names.
func DefinitionsForModel(modelName string) map[string]any {
	switch strings.ToLower(modelName) {
	case "user", "account", "member", "customer":
		return map[string]any{
			"name":  Lazy(func() any { return Name() }),
			"email": Lazy(func() any { return Email() }),
		}
	case "post", "article", "page", "story":
		return map[string]any{
			"title": Lazy(func() any { return Sentence() }),
			"body":  Lazy(func() any { return Paragraph() }),
		}
	case "comment", "review", "note":
		return map[string]any{
			"body":   Lazy(func() any { return Sentence() }),
			"author": Lazy(func() any { return Name() }),
		}
	default:
		return map[string]any{
			"name": Lazy(func() any { return Word() }),
		}
	}
}

// ForFieldType returns a lazy generator for a scaffold field type.
func ForFieldType(fieldType string) func() any {
	base, _, _ := strings.Cut(strings.ToLower(strings.TrimSpace(fieldType)), ":")
	switch base {
	case "email":
		return Lazy(func() any { return Email() })
	case "url":
		return Lazy(func() any { return URL() })
	case "text":
		return Lazy(func() any { return Paragraph() })
	case "int", "integer":
		return Lazy(func() any { return Int() })
	case "bigint", "references", "reference", "belongs_to":
		return Lazy(func() any { return Int64() })
	case "float", "decimal", "double":
		return Lazy(func() any { return Float() })
	case "bool", "boolean":
		return Lazy(func() any { return Bool() })
	case "datetime", "timestamp":
		return Lazy(func() any { return DateTime() })
	case "date":
		return Lazy(func() any { return Date() })
	case "time":
		return Lazy(func() any { return Time() })
	case "uuid":
		return Lazy(func() any { return UUID() })
	case "json", "jsonb":
		return Lazy(func() any { return "{}" })
	default:
		return Lazy(func() any { return Word() })
	}
}

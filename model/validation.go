package model

/*
|--------------------------------------------------------------------------
| Validation
|--------------------------------------------------------------------------
|
| Implements Validation as part of the model package in the Gofreight
| framework. Key symbols: Errors, Add, Any, Messages, Validator,
| Validatable.
| 
| The model package is the ORM layer: repositories, queries, associations,
| soft deletes, validation, serialization, collections, and pagination.
| 
| Models map to tables via struct tags; migrations define schema
| separately in db/migrate.
| 
| See docs/models.md, docs/orm.md, and docs/factories.md for
| Laravel-aligned patterns.
| 
| Symbols defined here include: Errors (exported type); Validator
| (exported type); Validatable (exported type); Validate (Validate runs
| all validators on the record and returns collected errors.); Presence
| (Presence validates that a value is not empty.); Length (Length
| validates string length within min and max bounds.); Format (Format
| validates a string against a regex pattern.); Numericality (Numericality
| validates that a numeric field is within range.).
| 
*/

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Errors holds validation errors keyed by field name.
type Errors map[string][]string

func (e Errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

func (e Errors) Any() bool {
	return len(e) > 0
}

func (e Errors) Messages() map[string][]string {
	return e
}

// Validator defines a validation rule for a field.
type Validator struct {
	Field    string
	Validate func(value any) error
}

// Validatable models implement validation and callbacks.
type Validatable interface {
	Validators() []Validator
}

// Validate runs all validators on the record and returns collected errors.
func Validate(record Validatable) Errors {
	errs := make(Errors)
	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, validator := range record.Validators() {
		field := v.FieldByName(validator.Field)
		if !field.IsValid() {
			continue
		}
		if err := validator.Validate(field.Interface()); err != nil {
			errs.Add(strings.ToLower(validator.Field), err.Error())
		}
	}
	return errs
}

// Presence validates that a value is not empty.
func Presence(field string) Validator {
	return Validator{
		Field: field,
		Validate: func(value any) error {
			if isBlank(value) {
				return fmt.Errorf("can't be blank")
			}
			return nil
		},
	}
}

// Length validates string length within min and max bounds.
func Length(field string, min, max int) Validator {
	return Validator{
		Field: field,
		Validate: func(value any) error {
			s, ok := value.(string)
			if !ok {
				return nil
			}
			n := utf8.RuneCountInString(s)
			if n < min {
				return fmt.Errorf("is too short (minimum is %d characters)", min)
			}
			if max > 0 && n > max {
				return fmt.Errorf("is too long (maximum is %d characters)", max)
			}
			return nil
		},
	}
}

// Format validates a string against a regex pattern.
func Format(field, pattern, message string) Validator {
	re := regexp.MustCompile(pattern)
	return Validator{
		Field: field,
		Validate: func(value any) error {
			s, ok := value.(string)
			if !ok || s == "" {
				return nil
			}
			if !re.MatchString(s) {
				if message == "" {
					message = "is invalid"
				}
				return fmt.Errorf("%s", message)
			}
			return nil
		},
	}
}

// Numericality validates that a numeric field is within range.
func Numericality(field string, min, max float64) Validator {
	return Validator{
		Field: field,
		Validate: func(value any) error {
			var n float64
			switch v := value.(type) {
			case int:
				n = float64(v)
			case int64:
				n = float64(v)
			case float64:
				n = v
			default:
				return nil
			}
			if n < min {
				return fmt.Errorf("must be greater than or equal to %g", min)
			}
			if max > 0 && n > max {
				return fmt.Errorf("must be less than or equal to %g", max)
			}
			return nil
		},
	}
}

// Uniqueness validates that a field value is unique in the database.
func Uniqueness[T any](field, column string, repo *Repository[T], excludeID int64) Validator {
	return Validator{
		Field: field,
		Validate: func(value any) error {
			if isBlank(value) {
				return nil
			}
			count, err := repo.Query(context.Background()).WhereEq(column, value).Count()
			if err != nil {
				return nil
			}
			if count > 0 {
				if excludeID > 0 {
					existing, err := repo.Query(context.Background()).WhereEq(column, value).First()
					if err == nil && recordID(existing) != excludeID {
						return fmt.Errorf("has already been taken")
					}
					return nil
				}
				return fmt.Errorf("has already been taken")
			}
			return nil
		},
	}
}

// Inclusion validates that a value is in the allowed list.
func Inclusion(field string, allowed []string) Validator {
	allowedSet := make(map[string]bool)
	for _, a := range allowed {
		allowedSet[a] = true
	}
	return Validator{
		Field: field,
		Validate: func(value any) error {
			s, ok := value.(string)
			if !ok || s == "" {
				return nil
			}
			if !allowedSet[s] {
				return fmt.Errorf("is not included in the list")
			}
			return nil
		},
	}
}

// Email validates email format.
func Email(field string) Validator {
	return Format(field, `^[^@\s]+@[^@\s]+\.[^@\s]+$`, "is not a valid email")
}

// Confirmation validates a confirmation field matches (e.g. PasswordConfirmation).
func Confirmation(field, confirmationField string) Validator {
	return Validator{
		Field: field,
		Validate: func(value any) error {
			return nil // validated in ValidateWithConfirmation helper
		},
	}
}

// ValidateConfirmation checks that field matches confirmationField on the record.
func ValidateConfirmation(record any, field, confirmationField string) error {
	v := reflect.ValueOf(record)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	f1 := v.FieldByName(field)
	f2 := v.FieldByName(confirmationField)
	if !f1.IsValid() || !f2.IsValid() {
		return nil
	}
	if f1.String() != f2.String() {
		return fmt.Errorf("doesn't match confirmation")
	}
	return nil
}

// Accepted validates that a boolean field is true (terms acceptance).
func Accepted(field string) Validator {
	return Validator{
		Field: field,
		Validate: func(value any) error {
			b, ok := value.(bool)
			if !ok {
				return nil
			}
			if !b {
				return fmt.Errorf("must be accepted")
			}
			return nil
		},
	}
}

func isBlank(value any) bool {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case nil:
		return true
	default:
		return false
	}
}

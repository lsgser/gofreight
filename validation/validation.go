package validation

/*
|--------------------------------------------------------------------------
| Validation
|--------------------------------------------------------------------------
|
| Implements Validation as part of the validation package in the Gofreight
| framework. Key symbols: Errors, Error, Validator, New, FromMap, Get.
| 
| Validation provides rule structs and runners used by models, form
| requests, and the vine DSL.
| 
| Errors map to field names for JSON and GFT #error directives.
| 
| Symbols defined here include: Errors (exported type); Validator
| (exported type); New (New creates a validator from form values.);
| FromMap (FromMap creates a validator from a map.); Get (Get returns a
| field value.); Required (Required ensures a field is present.); Email
| (Email validates email format.); MinLength (MinLength validates minimum
| string length.).
| 
*/

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Errors holds field validation errors.
type Errors map[string][]string

func (e Errors) Error() string {
	var parts []string
	for field, msgs := range e {
		parts = append(parts, fmt.Sprintf("%s: %s", field, strings.Join(msgs, ", ")))
	}
	return strings.Join(parts, "; ")
}

// Validator validates request input.
type Validator struct {
	data   map[string]string
	errors Errors
}

// New creates a validator from form values.
func New(r *http.Request) (*Validator, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	data := make(map[string]string)
	for k, v := range r.Form {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	return &Validator{data: data, errors: make(Errors)}, nil
}

// FromMap creates a validator from a map.
func FromMap(data map[string]string) *Validator {
	return &Validator{data: data, errors: make(Errors)}
}

// Get returns a field value.
func (v *Validator) Get(key string) string { return v.data[key] }

// Required ensures a field is present.
func (v *Validator) Required(fields ...string) *Validator {
	for _, f := range fields {
		if strings.TrimSpace(v.data[f]) == "" {
			v.add(f, "is required")
		}
	}
	return v
}

// Email validates email format.
func (v *Validator) Email(field string) *Validator {
	val := v.data[field]
	if val == "" {
		return v
	}
	re := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	if !re.MatchString(val) {
		v.add(field, "must be a valid email")
	}
	return v
}

// MinLength validates minimum string length.
func (v *Validator) MinLength(field string, min int) *Validator {
	if len(v.data[field]) < min {
		v.add(field, fmt.Sprintf("must be at least %d characters", min))
	}
	return v
}

// MaxLength validates maximum string length.
func (v *Validator) MaxLength(field string, max int) *Validator {
	if len(v.data[field]) > max {
		v.add(field, fmt.Sprintf("must be at most %d characters", max))
	}
	return v
}

// In validates value is in allowed set.
func (v *Validator) In(field string, allowed ...string) *Validator {
	val := v.data[field]
	if val == "" {
		return v
	}
	for _, a := range allowed {
		if val == a {
			return v
		}
	}
	v.add(field, "has an invalid value")
	return v
}

// Int converts and validates integer field.
func (v *Validator) Int(field string) (int, bool) {
	n, err := strconv.Atoi(v.data[field])
	if err != nil {
		v.add(field, "must be an integer")
		return 0, false
	}
	return n, true
}

// Fails returns true if validation failed.
func (v *Validator) Fails() bool { return len(v.errors) > 0 }

// Errors returns validation errors.
func (v *Validator) Errors() Errors { return v.errors }

func (v *Validator) add(field, msg string) {
	v.errors[field] = append(v.errors[field], msg)
}

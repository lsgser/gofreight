package request

/*
|--------------------------------------------------------------------------
| Form
|--------------------------------------------------------------------------
|
| Implements Form as part of the request package in the Gofreight
| framework. Key symbols: FormRequest, Rule, NewFormRequest, FromJSON,
| Authorize, Required.
| 
| Form requests bind and validate HTTP input using vine schemas or
| validation tags before controller actions run.
| 
| Integrates with controller Base for 422 Unprocessable responses and GFT
| error display helpers.
| 
| Symbols defined here include: FormRequest (exported type); Rule
| (exported type); NewFormRequest (NewFormRequest creates a form request
| from an HTTP request.); FromJSON (FromJSON creates a form request from
| JSON body.); Authorize (Authorize sets whether the request is
| authorized.); Required (Required adds required rule.); Email (Email adds
| email validation.); MinLength (MinLength adds minimum length
| validation.).
| 
*/

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/lsgser/gofreight/validation"
)

// FormRequest validates incoming HTTP requests with chainable rules.
type FormRequest struct {
	rules    []Rule
	data     map[string]string
	Errors   validation.Errors
	Authorized bool
}

// Rule validates a single field.
type Rule struct {
	Field string
	Check func(value string) string // empty = ok, else error message
}

// NewFormRequest creates a form request from an HTTP request.
func NewFormRequest(r *http.Request) (*FormRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	data := make(map[string]string)
	for k, v := range r.Form {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	return &FormRequest{data: data, Authorized: true, Errors: make(validation.Errors)}, nil
}

// FromJSON creates a form request from JSON body.
func FromJSON(r *http.Request) (*FormRequest, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]string
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return &FormRequest{data: data, Authorized: true, Errors: make(validation.Errors)}, nil
}

// Authorize sets whether the request is authorized.
func (f *FormRequest) Authorize(ok bool) *FormRequest {
	f.Authorized = ok
	return f
}

// Required adds required rule.
func (f *FormRequest) Required(fields ...string) *FormRequest {
	for _, field := range fields {
		fld := field
		f.rules = append(f.rules, Rule{Field: fld, Check: func(v string) string {
			if v == "" {
				return "is required"
			}
			return ""
		}})
	}
	return f
}

// Email adds email validation.
func (f *FormRequest) Email(field string) *FormRequest {
	v := validation.FromMap(f.data)
	v.Email(field)
	f.mergeErrors(v.Errors())
	return f
}

// MinLength adds minimum length validation.
func (f *FormRequest) MinLength(field string, n int) *FormRequest {
	f.rules = append(f.rules, Rule{Field: field, Check: func(v string) string {
		if len(v) < n {
			return "is too short"
		}
		return ""
	}})
	return f
}

// Validate runs all rules.
func (f *FormRequest) Validate() bool {
	for _, rule := range f.rules {
		if msg := rule.Check(f.data[rule.Field]); msg != "" {
			f.Errors[rule.Field] = append(f.Errors[rule.Field], msg)
		}
	}
	return len(f.Errors) == 0 && f.Authorized
}

func (f *FormRequest) mergeErrors(e validation.Errors) {
	for k, v := range e {
		f.Errors[k] = append(f.Errors[k], v...)
	}
}

// Get returns a field value.
func (f *FormRequest) Get(key string) string { return f.data[key] }

// All returns all validated data.
func (f *FormRequest) All() map[string]string { return f.data }

// Failed returns true if validation failed.
func (f *FormRequest) Failed() bool { return !f.Authorized || len(f.Errors) > 0 }

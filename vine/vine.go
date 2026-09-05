package vine

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lsgser/gofreight/validation"
)

// Rule validates a single field value.
type Rule interface {
	Validate(field, value string) []string
}

// ObjectSchema describes validation rules for a set of fields.
type ObjectSchema struct {
	fields map[string]Rule
}

// Object builds a schema from field rules.
func Object(fields map[string]Rule) *ObjectSchema {
	return &ObjectSchema{fields: fields}
}

// Compile returns the schema (mirrors schema-first validator APIs).
func Compile(schema *ObjectSchema) *ObjectSchema {
	return schema
}

// Validate checks data against the schema and returns cleaned data + errors.
func (s *ObjectSchema) Validate(data map[string]string) (map[string]string, validation.Errors) {
	errs := make(validation.Errors)
	out := make(map[string]string, len(data))
	for k, v := range data {
		out[k] = v
	}
	for field, rule := range s.fields {
		msgs := rule.Validate(field, data[field])
		if len(msgs) > 0 {
			errs[field] = append(errs[field], msgs...)
		}
	}
	return out, errs
}

type stringRule struct {
	checks []func(string) string
}

// String starts a string field rule chain.
func String() *stringRule {
	return &stringRule{}
}

func (s *stringRule) Required() *stringRule {
	s.checks = append(s.checks, func(v string) string {
		if strings.TrimSpace(v) == "" {
			return "is required"
		}
		return ""
	})
	return s
}

func (s *stringRule) Email() *stringRule {
	re := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	s.checks = append(s.checks, func(v string) string {
		if v == "" {
			return ""
		}
		if !re.MatchString(v) {
			return "must be a valid email"
		}
		return ""
	})
	return s
}

func (s *stringRule) MinLength(n int) *stringRule {
	s.checks = append(s.checks, func(v string) string {
		if len(v) < n {
			return fmt.Sprintf("must be at least %d characters", n)
		}
		return ""
	})
	return s
}

func (s *stringRule) MaxLength(n int) *stringRule {
	s.checks = append(s.checks, func(v string) string {
		if len(v) > n {
			return fmt.Sprintf("must be at most %d characters", n)
		}
		return ""
	})
	return s
}

func (s *stringRule) In(values ...string) *stringRule {
	allowed := append([]string{}, values...)
	s.checks = append(s.checks, func(v string) string {
		if v == "" {
			return ""
		}
		for _, a := range allowed {
			if v == a {
				return ""
			}
		}
		return "has an invalid value"
	})
	return s
}

func (s *stringRule) Validate(_ string, value string) []string {
	return runChecks(value, s.checks)
}

type boolRule struct {
	checks []func(string) string
}

// Boolean validates checkbox-style fields ("1", "true", "on").
func Boolean() *boolRule {
	return &boolRule{}
}

func (b *boolRule) Required() *boolRule {
	b.checks = append(b.checks, func(v string) string {
		if !isTruthy(v) {
			return "must be accepted"
		}
		return ""
	})
	return b
}

func (b *boolRule) Validate(_ string, value string) []string {
	return runChecks(value, b.checks)
}

type numberRule struct {
	checks []func(string) string
}

// Number validates numeric string fields.
func Number() *numberRule {
	return &numberRule{}
}

func (n *numberRule) Required() *numberRule {
	n.checks = append(n.checks, func(v string) string {
		if strings.TrimSpace(v) == "" {
			return "is required"
		}
		return ""
	})
	return n
}

func (n *numberRule) Validate(_ string, value string) []string {
	n.checks = append([]func(string) string{func(v string) string {
		if v == "" {
			return ""
		}
		for _, ch := range v {
			if ch < '0' || ch > '9' {
				if ch != '-' || len(v) == 1 {
					return "must be a number"
				}
			}
		}
		return ""
	}}, n.checks...)
	return runChecks(value, n.checks)
}

func runChecks(value string, checks []func(string) string) []string {
	var msgs []string
	for _, check := range checks {
		if msg := check(value); msg != "" {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "on", "yes":
		return true
	default:
		return false
	}
}

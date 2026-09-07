package model

import (
	"encoding/json"
	"reflect"
	"strings"
)

// AppendFunc computes an appended JSON attribute for a model.
type AppendFunc func(record any) any

// Serializer converts models to arrays/JSON (Laravel toArray / toJson).
type Serializer struct {
	record        any
	hidden        map[string]struct{}
	visible       map[string]struct{}
	useHidden     bool
	useVisible    bool
	appends       map[string]AppendFunc
	useAppends    bool
	skipAppends   bool
	includeAssocs bool
	dateFormat    string
}

// Serialize begins serialization for a model or struct pointer.
func Serialize(record any) *Serializer {
	return &Serializer{
		record:     record,
		hidden:     make(map[string]struct{}),
		visible:    make(map[string]struct{}),
		appends:    make(map[string]AppendFunc),
		useAppends: true,
	}
}

// Hidden marks attributes hidden from output (Laravel $hidden).
func (s *Serializer) Hidden(keys ...string) *Serializer {
	s.useHidden = true
	for _, k := range keys {
		s.hidden[strings.ToLower(k)] = struct{}{}
	}
	return s
}

// Visible sets an allow-list of attributes (Laravel $visible).
func (s *Serializer) Visible(keys ...string) *Serializer {
	s.useVisible = true
	s.visible = make(map[string]struct{})
	for _, k := range keys {
		s.visible[strings.ToLower(k)] = struct{}{}
	}
	return s
}

// MakeVisible temporarily exposes hidden attributes.
func (s *Serializer) MakeVisible(keys ...string) *Serializer {
	for _, k := range keys {
		delete(s.hidden, strings.ToLower(k))
	}
	s.useHidden = len(s.hidden) > 0
	return s
}

// MakeHidden temporarily hides attributes.
func (s *Serializer) MakeHidden(keys ...string) *Serializer {
	s.useHidden = true
	for _, k := range keys {
		s.hidden[strings.ToLower(k)] = struct{}{}
	}
	return s
}

// SetVisible replaces the visible attribute list.
func (s *Serializer) SetVisible(keys ...string) *Serializer {
	return s.Visible(keys...)
}

// SetHidden replaces the hidden attribute list.
func (s *Serializer) SetHidden(keys ...string) *Serializer {
	s.hidden = make(map[string]struct{})
	s.useHidden = true
	for _, k := range keys {
		s.hidden[strings.ToLower(k)] = struct{}{}
	}
	return s
}

// Append registers a computed attribute (Laravel $appends / append()).
func (s *Serializer) Append(key string, fn AppendFunc) *Serializer {
	s.appends[strings.ToLower(key)] = fn
	return s
}

// SetAppends replaces appended attributes.
func (s *Serializer) SetAppends(appends map[string]AppendFunc) *Serializer {
	s.appends = make(map[string]AppendFunc)
	for k, fn := range appends {
		s.appends[strings.ToLower(k)] = fn
	}
	s.useAppends = true
	return s
}

// WithoutAppends excludes appended attributes.
func (s *Serializer) WithoutAppends() *Serializer {
	s.skipAppends = true
	return s
}

// WithAssociations includes eager-loaded association data in output.
func (s *Serializer) WithAssociations() *Serializer {
	s.includeAssocs = true
	return s
}

// DateFormat sets the format for timestamp fields (Laravel serializeDate).
func (s *Serializer) DateFormat(layout string) *Serializer {
	s.dateFormat = layout
	return s
}

// ToArray converts the model to a map (Laravel toArray).
func (s *Serializer) ToArray() map[string]any {
	return s.buildMap()
}

// AttributesToArray converts attributes only (no associations).
func (s *Serializer) AttributesToArray() map[string]any {
	prev := s.includeAssocs
	s.includeAssocs = false
	out := s.buildMap()
	s.includeAssocs = prev
	return out
}

// ToJSON converts the model to JSON bytes (Laravel toJson).
func (s *Serializer) ToJSON() ([]byte, error) {
	return json.Marshal(s.ToArray())
}

// ToMap is an alias for ToArray (api.Resource compatibility).
func (s *Serializer) ToMap() map[string]any {
	return s.ToArray()
}

func (s *Serializer) buildMap() map[string]any {
	if s.record == nil {
		return map[string]any{}
	}
	v := reflect.ValueOf(s.record)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return map[string]any{}
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return map[string]any{"value": v.Interface()}
	}

	result := make(map[string]any)
	collectSerializeFields(v, v.Type(), result, s)

	if s.includeAssocs {
		for k, v := range AssociationData(s.record) {
			result[toSnakeKey(k)] = v
		}
	}

	if s.useAppends && !s.skipAppends {
		for key, fn := range s.appends {
			if fn != nil {
				result[key] = fn(s.record)
			}
		}
	}
	return result
}

func collectSerializeFields(v reflect.Value, t reflect.Type, result map[string]any, s *Serializer) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		if field.Anonymous {
			collectSerializeFields(v.Field(i), field.Type, result, s)
			continue
		}

		key, skip := serializeKey(field)
		if skip || key == "" {
			continue
		}
		lowerKey := strings.ToLower(key)

		if s.useVisible {
			if _, ok := s.visible[lowerKey]; !ok {
				continue
			}
		}
		if s.useHidden {
			if _, ok := s.hidden[lowerKey]; ok {
				continue
			}
		}
		if field.Tag.Get("json") == "-" {
			continue
		}

		val := v.Field(i).Interface()
		result[key] = val
	}
}

func serializeKey(field reflect.StructField) (string, bool) {
	if tag := field.Tag.Get("json"); tag != "" {
		if tag == "-" {
			return "", true
		}
		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			return name, false
		}
	}
	if tag := field.Tag.Get("db"); tag != "" {
		if tag == "-" {
			return "", true
		}
		return tag, false
	}
	return toSnakeKey(field.Name), false
}

func toSnakeKey(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ToArray converts any model to a map using json/db tags.
func ToArray(record any) map[string]any {
	return Serialize(record).ToArray()
}

// ToJSON converts any model to JSON.
func ToJSON(record any) ([]byte, error) {
	return Serialize(record).ToJSON()
}

// AttributesToArray converts model attributes without associations.
func AttributesToArray(record any) map[string]any {
	return Serialize(record).AttributesToArray()
}

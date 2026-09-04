package gftest

import (
	"reflect"
	"strings"
)

func setAttrs(record any, attrs map[string]any) {
	v := reflect.ValueOf(record).Elem()
	t := v.Type()

	for key, val := range attrs {
		goName := snakeToCamel(key)
		setFieldRecursive(v, t, goName, val)
	}
}

func setFieldRecursive(v reflect.Value, t reflect.Type, goName string, val any) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			setFieldRecursive(v.Field(i), field.Type, goName, val)
			continue
		}
		dbTag := field.Tag.Get("db")
		if field.Name == goName || dbTag == keyOrSnake(goName) {
			f := v.Field(i)
			if f.IsValid() && f.CanSet() {
				rv := reflect.ValueOf(val)
				if rv.Type().AssignableTo(f.Type()) {
					f.Set(rv)
				} else if rv.Type().ConvertibleTo(f.Type()) {
					f.Set(rv.Convert(f.Type()))
				}
			}
			return
		}
	}
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func keyOrSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

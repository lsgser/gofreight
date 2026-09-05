package model

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/lsgser/gofreight/database"
)

// columnMeta describes a struct field mapped to a database column.
type columnMeta struct {
	Name     string
	DBName   string
	GoName   string
	Kind     reflect.Kind
	Index    int
	ReadOnly bool
}

func tableColumns[T any]() []columnMeta {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var cols []columnMeta
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() && field.Name[0] < 'A' || field.Name[0] > 'Z' {
			// skip unexported except embedded Record fields
			if field.Anonymous {
				embedded := tableColumnsFromType(field.Type)
				cols = append(cols, embedded...)
			}
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			if field.Anonymous {
				cols = append(cols, tableColumnsFromType(field.Type)...)
				continue
			}
			continue
		}
		cols = append(cols, columnMeta{
			Name:   dbTag,
			DBName: dbTag,
			GoName: field.Name,
			Kind:   field.Type.Kind(),
			Index:  i,
		})
	}
	return cols
}

func tableColumnsFromType(t reflect.Type) []columnMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var cols []columnMeta
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		cols = append(cols, columnMeta{
			Name:   dbTag,
			DBName: dbTag,
			GoName: field.Name,
			Kind:   field.Type.Kind(),
			Index:  i,
		})
	}
	return cols
}

func goFieldNameForColumn[T any](column string) string {
	for _, c := range tableColumns[T]() {
		if c.DBName == column || strings.EqualFold(c.GoName, column) {
			return c.GoName
		}
	}
	return column
}

func scanRows[T any](rows *sql.Rows) ([]T, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []T
	for rows.Next() {
		var record T
		dest := makeScanDest(&record, columns)
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		results = append(results, record)
	}
	return results, rows.Err()
}

func scanInto(row *sql.Row, dest any) error {
	v := reflect.ValueOf(dest).Elem()

	idField := v.FieldByName("ID")
	createdField := v.FieldByName("CreatedAt")
	updatedField := v.FieldByName("UpdatedAt")

	scanTargets := []any{idField.Addr().Interface()}
	if createdField.IsValid() {
		scanTargets = append(scanTargets, createdField.Addr().Interface())
	}
	if updatedField.IsValid() {
		scanTargets = append(scanTargets, updatedField.Addr().Interface())
	}

	return row.Scan(scanTargets...)
}

func makeScanDest(record any, columns []string) []any {
	v := reflect.ValueOf(record).Elem()
	t := v.Type()
	dest := make([]any, len(columns))

	for i, col := range columns {
		found := false
		for j := 0; j < t.NumField(); j++ {
			field := t.Field(j)
			if field.Anonymous {
				embedded := reflect.New(field.Type).Interface()
				embDest := makeScanDest(embedded, []string{col})
				if embDest[0] != nil {
					// copy embedded field address
					embVal := v.Field(j)
					if embVal.Kind() == reflect.Struct {
						for k := 0; k < field.Type.NumField(); k++ {
							ef := field.Type.Field(k)
							if ef.Tag.Get("db") == col {
								dest[i] = embVal.Field(k).Addr().Interface()
								found = true
								break
							}
						}
					}
				}
				if found {
					break
				}
			}
			dbTag := field.Tag.Get("db")
			if dbTag == col || strings.EqualFold(field.Name, col) {
				dest[i] = v.Field(j).Addr().Interface()
				found = true
				break
			}
		}
		if !found {
			dest[i] = new(any)
		}
	}
	return dest
}

func extractFields(record any, skipID bool) (columns []string, placeholders []string, values []any) {
	v := reflect.ValueOf(record).Elem()
	t := v.Type()
	collectFields(v, t, skipID, &columns, &placeholders, &values)
	return
}

func collectFields(v reflect.Value, t reflect.Type, skipID bool, columns *[]string, placeholders *[]string, values *[]any) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			collectFields(v.Field(i), field.Type, skipID, columns, placeholders, values)
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		if dbTag == "id" && skipID {
			continue
		}
		if !skipID && (dbTag == "created_at" || dbTag == "updated_at") {
			continue
		}

		fv := v.Field(i)
		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			continue
		}

		*columns = append(*columns, dbTag)
		*placeholders = append(*placeholders, database.Placeholder(len(*values)+1))
		*values = append(*values, fv.Interface())
	}
}

func setField(record any, name string, value int64) {
	v := reflect.ValueOf(record).Elem()
	setFieldRecursive(v, v.Type(), name, value)
}

func setFieldRecursive(v reflect.Value, t reflect.Type, name string, value int64) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			setFieldRecursive(v.Field(i), field.Type, name, value)
			continue
		}
		if field.Name == name {
			f := v.Field(i)
			if f.IsValid() && f.CanSet() {
				f.SetInt(value)
			}
			return
		}
	}
}

func recordID(record any) int64 {
	v := reflect.ValueOf(record).Elem()
	return fieldInt(v, "ID")
}

func fieldInt(v reflect.Value, name string) int64 {
	f := v.FieldByName(name)
	if f.IsValid() {
		return f.Int()
	}
	return 0
}

func setFieldValue(record any, goName string, value any) {
	v := reflect.ValueOf(record).Elem()
	f := v.FieldByName(goName)
	if f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(value))
	}
}

func newRecord[T any]() *T {
	var r T
	return &r
}

func cloneRecord[T any](src T) T {
	// shallow copy is sufficient for structs
	return src
}

func applyAttributes[T any](record *T, attrs map[string]any) {
	v := reflect.ValueOf(record).Elem()
	for k, val := range attrs {
		goName := goFieldNameForColumn[T](k)
		f := v.FieldByName(goName)
		if f.IsValid() && f.CanSet() {
			rv := reflect.ValueOf(val)
			if rv.Type().AssignableTo(f.Type()) {
				f.Set(rv)
			}
		}
	}
}

func recordToMap(record any) map[string]any {
	v := reflect.ValueOf(record).Elem()
	t := v.Type()
	result := make(map[string]any)
	collectMap(v, t, result)
	return result
}

func collectMap(v reflect.Value, t reflect.Type, result map[string]any) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			collectMap(v.Field(i), field.Type, result)
			continue
		}
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}
		result[dbTag] = v.Field(i).Interface()
	}
}

func quoteIdent(name string) string {
	// basic SQL injection guard for identifiers
	if strings.ContainsAny(name, " ;'\"") {
		panic(fmt.Sprintf("invalid SQL identifier: %s", name))
	}
	return name
}

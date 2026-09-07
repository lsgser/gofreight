package model

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/lsgser/gofreight/database"
)

// AssocType defines association relationship kinds.
type AssocType string

const (
	BelongsTo  AssocType = "belongs_to"
	HasOne     AssocType = "has_one"
	HasMany    AssocType = "has_many"
	ManyToMany AssocType = "many_to_many"
)

// Association defines a relationship between models.
type Association struct {
	Name         string
	Type         AssocType
	ForeignKey   string
	ReferenceKey string
	Table        string
	PivotTable   string
	LocalKey     string
	ForeignPivot string
	RelatedPivot string
}

// BelongsToAssociation creates a belongs_to association definition.
func BelongsToAssociation(name, table, foreignKey string) Association {
	return Association{
		Name:         name,
		Type:         BelongsTo,
		Table:        table,
		ForeignKey:   foreignKey,
		ReferenceKey: "id",
	}
}

// HasManyAssociation creates a has_many association definition.
func HasManyAssociation(name, table, foreignKey string) Association {
	return Association{
		Name:         name,
		Type:         HasMany,
		Table:        table,
		ForeignKey:   foreignKey,
		ReferenceKey: "id",
		LocalKey:     "id",
	}
}

// HasOneAssociation creates a has_one association definition.
func HasOneAssociation(name, table, foreignKey string) Association {
	return Association{
		Name:       name,
		Type:       HasOne,
		Table:      table,
		ForeignKey: foreignKey,
		LocalKey:   "id",
	}
}

// ManyToManyAssociation creates a many_to_many association definition.
func ManyToManyAssociation(name, table, pivot, localPivot, foreignPivot string) Association {
	return Association{
		Name:         name,
		Type:         ManyToMany,
		Table:        table,
		PivotTable:   pivot,
		LocalKey:     "id",
		ForeignPivot: localPivot,
		RelatedPivot: foreignPivot,
	}
}

func (r *Repository[T]) preloadAssociations(ctx context.Context, records []T, names []string) error {
	for _, name := range names {
		assoc, ok := r.associations[name]
		if !ok {
			continue
		}
		if err := r.loadAssociation(ctx, records, assoc); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository[T]) loadAssociation(ctx context.Context, records []T, assoc Association) error {
	if len(records) == 0 {
		return nil
	}

	switch assoc.Type {
	case HasMany:
		return r.loadHasMany(ctx, records, assoc)
	case BelongsTo:
		return r.loadBelongsTo(ctx, records, assoc)
	case HasOne:
		return r.loadHasOne(ctx, records, assoc)
	case ManyToMany:
		return r.loadManyToMany(ctx, records, assoc)
	}
	return nil
}

func (r *Repository[T]) loadHasMany(ctx context.Context, records []T, assoc Association) error {
	ids := make([]any, len(records))
	for i, rec := range records {
		ids[i] = recordID(&rec)
	}

	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s IN (%s)",
		assoc.Table, assoc.ForeignKey, placeholders(len(ids)))

	rows, err := database.DB().QueryContext(ctx, sql, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	relatedByFK := make(map[int64][]map[string]any)
	columns, _ := rows.Columns()
	for rows.Next() {
		vals := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		row := make(map[string]any)
		var fk int64
		for i, col := range columns {
			row[col] = vals[i]
			if col == assoc.ForeignKey {
				fk = toInt64(vals[i])
			}
		}
		relatedByFK[fk] = append(relatedByFK[fk], row)
	}

	for i, rec := range records {
		id := recordID(&rec)
		if related, ok := relatedByFK[id]; ok {
			setAssociationData(&records[i], assoc.Name, related)
		}
		_ = rec
	}
	return nil
}

func (r *Repository[T]) loadBelongsTo(ctx context.Context, records []T, assoc Association) error {
	fkValues := make([]any, 0, len(records))
	fkSet := make(map[int64]bool)
	for _, rec := range records {
		fk := foreignKeyValue(&rec, assoc.ForeignKey)
		if fk > 0 && !fkSet[fk] {
			fkValues = append(fkValues, fk)
			fkSet[fk] = true
		}
	}
	if len(fkValues) == 0 {
		return nil
	}

	sql := fmt.Sprintf("SELECT * FROM %s WHERE id IN (%s)", assoc.Table, placeholders(len(fkValues)))
	rows, err := database.DB().QueryContext(ctx, sql, fkValues...)
	if err != nil {
		return err
	}
	defer rows.Close()

	relatedByID := make(map[int64]map[string]any)
	columns, _ := rows.Columns()
	for rows.Next() {
		vals := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		row := make(map[string]any)
		var id int64
		for i, col := range columns {
			row[col] = vals[i]
			if col == "id" {
				id = toInt64(vals[i])
			}
		}
		relatedByID[id] = row
	}

	for i, rec := range records {
		fk := foreignKeyValue(&rec, assoc.ForeignKey)
		if related, ok := relatedByID[fk]; ok {
			setAssociationData(&records[i], assoc.Name, related)
		}
	}
	return nil
}

func (r *Repository[T]) loadHasOne(ctx context.Context, records []T, assoc Association) error {
	return r.loadHasMany(ctx, records, Association{
		Name: assoc.Name, Type: HasOne, Table: assoc.Table, ForeignKey: assoc.ForeignKey,
	})
}

func (r *Repository[T]) loadManyToMany(ctx context.Context, records []T, assoc Association) error {
	ids := make([]any, len(records))
	for i, rec := range records {
		ids[i] = recordID(&rec)
	}

	sql := fmt.Sprintf(
		"SELECT %s.*, %s.%s AS _pivot_local FROM %s INNER JOIN %s ON %s.id = %s.%s WHERE %s.%s IN (%s)",
		assoc.Table, assoc.PivotTable, assoc.ForeignPivot,
		assoc.Table, assoc.PivotTable, assoc.Table, assoc.PivotTable, assoc.RelatedPivot,
		assoc.PivotTable, assoc.ForeignPivot, placeholders(len(ids)),
	)

	rows, err := database.DB().QueryContext(ctx, sql, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	relatedByLocal := make(map[int64][]map[string]any)
	columns, _ := rows.Columns()
	for rows.Next() {
		vals := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		row := make(map[string]any)
		var localID int64
		for i, col := range columns {
			if col == "_pivot_local" {
				localID = toInt64(vals[i])
			} else {
				row[col] = vals[i]
			}
		}
		relatedByLocal[localID] = append(relatedByLocal[localID], row)
	}

	for i, rec := range records {
		id := recordID(&rec)
		if related, ok := relatedByLocal[id]; ok {
			setAssociationData(&records[i], assoc.Name, related)
		}
		_ = rec
	}
	return nil
}

func foreignKeyValue(record any, fkColumn string) int64 {
	m := recordToMap(record)
	if v, ok := m[fkColumn]; ok {
		return toInt64(v)
	}
	goName := snakeToTitle(fkColumn)
	v := reflect.ValueOf(record).Elem().FieldByName(goName)
	if v.IsValid() {
		return v.Int()
	}
	return 0
}

func setAssociationData(record any, name string, data any) {
	assocCache.Store(record, name, data)
}

var assocCache = &associationCache{data: make(map[uintptr]map[string]any)}

type associationCache struct {
	data map[uintptr]map[string]any
}

func (c *associationCache) Store(record any, name string, data any) {
	ptr := reflect.ValueOf(record).Pointer()
	if c.data[ptr] == nil {
		c.data[ptr] = make(map[string]any)
	}
	c.data[ptr][name] = data
}

// GetAssociation retrieves preloaded association data for a record.
func GetAssociation(record any, name string) (any, bool) {
	ptr := reflect.ValueOf(record).Pointer()
	if m, ok := assocCache.data[ptr]; ok {
		v, ok := m[name]
		return v, ok
	}
	return nil, false
}

// AssociationData returns all eager-loaded associations for serialization.
func AssociationData(record any) map[string]any {
	ptr := reflect.ValueOf(record).Pointer()
	if m, ok := assocCache.data[ptr]; ok {
		out := make(map[string]any, len(m))
		for k, v := range m {
			out[k] = v
		}
		return out
	}
	return nil
}

func placeholders(n int) string {
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = database.Placeholder(i + 1)
	}
	return strings.Join(parts, ", ")
}

func snakeToTitle(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}

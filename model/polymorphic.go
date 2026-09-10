package model

/*
|--------------------------------------------------------------------------
| Polymorphic
|--------------------------------------------------------------------------
|
| Implements Polymorphic as part of the model package in the Gofreight
| framework. Key symbols: PolymorphicAssoc, PolymorphicBelongsTo,
| PolymorphicHasMany, NestedSave.
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
| Symbols defined here include: PolymorphicAssoc (exported type);
| PolymorphicBelongsTo (PolymorphicBelongsTo loads the parent record for a
| polymorphic association.); PolymorphicHasMany (PolymorphicHasMany loads
| child records for a polymorphic association.); NestedSave
| (NestedAttributes saves associated records (pass pre-built records via
| callback).).
| 
*/

import (
	"context"
	"fmt"
	"strings"

	"github.com/lsgser/gofreight/database"
)

// PolymorphicAssoc defines a polymorphic belongs_to/has_many relationship.
type PolymorphicAssoc struct {
	Name      string
	TypeField string
	IDField   string
	Table     string
}

// PolymorphicBelongsTo loads the parent record for a polymorphic association.
func PolymorphicBelongsTo(ctx context.Context, typeName string, id any) (map[string]any, error) {
	table := pluralizeTable(typeName)
	q := fmt.Sprintf("SELECT * FROM %s WHERE id = %s LIMIT 1", table, database.Placeholder(1))
	row := database.DB().QueryRowContext(ctx, q, id)
	cols, err := columnNames(ctx, table)
	if err != nil {
		return nil, err
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := row.Scan(ptrs...); err != nil {
		return nil, err
	}
	result := make(map[string]any, len(cols))
	for i, c := range cols {
		result[c] = vals[i]
	}
	return result, nil
}

// PolymorphicHasMany loads child records for a polymorphic association.
func PolymorphicHasMany(ctx context.Context, assoc PolymorphicAssoc, typeName string, id any) ([]map[string]any, error) {
	q := fmt.Sprintf("SELECT * FROM %s WHERE %s = %s AND %s = %s",
		assoc.Table, assoc.TypeField, database.Placeholder(1), assoc.IDField, database.Placeholder(2))
	rows, err := database.DB().QueryContext(ctx, q, typeName, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRowMaps(rows)
}

// NestedAttributes saves associated records (pass pre-built records via callback).
func NestedSave[T any](ctx context.Context, repo *Repository[T], items []*T) error {
	for _, item := range items {
		if err := repo.Create(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func pluralizeTable(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, "y") {
		return strings.TrimSuffix(lower, "y") + "ies"
	}
	return lower + "s"
}

func columnNames(ctx context.Context, table string) ([]string, error) {
	rows, err := database.DB().QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s LIMIT 0", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rows.Columns()
}

func scanRowMaps(rows interface {
	Next() bool
	Scan(...any) error
	Columns() ([]string, error)
}) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var results []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			row[c] = vals[i]
		}
		results = append(results, row)
	}
	return results, nil
}

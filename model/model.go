package model

/*
|--------------------------------------------------------------------------
| Model
|--------------------------------------------------------------------------
|
| Implements Model as part of the model package in the Gofreight
| framework. Key symbols: Record, IsSoftDeleted, Repository,
| NewRepository, Query, Association.
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
| Symbols defined here include: Record (exported type); IsSoftDeleted
| (IsSoftDeleted reports whether the record has a deleted_at timestamp
| set.); Repository (exported type); NewRepository (NewRepository creates
| an ORM repository for the given table.); Query (Query returns a
| chainable query builder for this model.); Association (Association
| registers a model association.); Scope (Scope registers a named,
| reusable query fragment on the repository.); EnableSoftDelete
| (EnableSoftDelete enables soft delete for this model.).
| 
*/

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/lsgser/gofreight/database"
)

// Record is the base struct for all models with id and timestamps.
type Record struct {
	ID        int64   `db:"id" json:"id"`
	CreatedAt string  `db:"created_at" json:"created_at"`
	UpdatedAt string  `db:"updated_at" json:"updated_at"`
	DeletedAt *string `db:"deleted_at" json:"deleted_at,omitempty"`
}

// IsSoftDeleted reports whether the record has a deleted_at timestamp set.
func (r *Record) IsSoftDeleted() bool {
	return r.DeletedAt != nil && *r.DeletedAt != ""
}

// Repository is the ORM entry point for a model table.
type Repository[T any] struct {
	TableName     string
	associations  map[string]Association
	scopes        map[string]ScopeFunc[T]
	softDelete    bool
	deletedColumn string
}

// NewRepository creates an ORM repository for the given table.
func NewRepository[T any](table string) *Repository[T] {
	return &Repository[T]{
		TableName:    table,
		associations: make(map[string]Association),
		scopes:       make(map[string]ScopeFunc[T]),
	}
}

// Query returns a chainable query builder for this model.
func (r *Repository[T]) Query(ctx context.Context) *Query[T] {
	q := NewQuery[T](r.TableName, ctx)
	q.repo = r
	q.scopeFuncs = r.scopes
	if r.softDelete {
		q.SoftDelete(r.deletedColumnName())
	}
	return q
}

// Association registers a model association.
func (r *Repository[T]) Association(assoc Association) *Repository[T] {
	r.associations[assoc.Name] = assoc
	return r
}

// Scope registers a named, reusable query fragment on the repository.
func (r *Repository[T]) Scope(name string, fn ScopeFunc[T]) *Repository[T] {
	r.scopes[name] = fn
	return r
}

// EnableSoftDelete enables soft delete for this model.
// Add a nullable deleted_at column to your table (see docs/orm.md).
func (r *Repository[T]) EnableSoftDelete(column ...string) *Repository[T] {
	r.softDelete = true
	r.deletedColumn = "deleted_at"
	if len(column) > 0 && column[0] != "" {
		r.deletedColumn = column[0]
	}
	return r
}

// --- Convenience methods delegating to Query ---

// All returns all records.
func (r *Repository[T]) All(ctx context.Context) ([]T, error) {
	return r.Query(ctx).Get()
}

// Find retrieves a record by ID.
func (r *Repository[T]) Find(ctx context.Context, id int64) (*T, error) {
	return r.Query(ctx).Find(id)
}

// FindBy retrieves the first record matching a column.
func (r *Repository[T]) FindBy(ctx context.Context, column string, value any) (*T, error) {
	return r.Query(ctx).FindBy(column, value)
}

// First returns the first record.
func (r *Repository[T]) First(ctx context.Context) (*T, error) {
	return r.Query(ctx).First()
}

// Where finds records matching a column value (backward compatible).
func (r *Repository[T]) Where(ctx context.Context, column string, value any) ([]T, error) {
	return r.Query(ctx).WhereEq(column, value).Get()
}

// Count returns total record count.
func (r *Repository[T]) Count(ctx context.Context) (int64, error) {
	return r.Query(ctx).Count()
}

// Paginate returns a paginated result set.
func (r *Repository[T]) Paginate(ctx context.Context, page, perPage int) (*Page[T], error) {
	return r.Query(ctx).Paginate(page, perPage)
}

// Create inserts a new record with callbacks.
func (r *Repository[T]) Create(ctx context.Context, record *T) error {
	if err := RunCallbacks(ctx, record, BeforeCreate); err != nil {
		return err
	}
	if err := RunCallbacks(ctx, record, BeforeSave); err != nil {
		return err
	}

	columns, placeholders, values := extractFields(record, true)
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)%s",
		r.TableName,
		joinStrings(columns),
		joinStrings(placeholders),
		database.ReturningClause("id", "created_at", "updated_at"),
	)

	if database.ReturningClause("id") == "" {
		result, err := execContext(ctx, query, values...)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err == nil {
			setField(record, "ID", id)
		}
	} else {
		row := queryRowContext(ctx, query, values...)
		if err := scanInto(row, record); err != nil {
			return err
		}
	}

	if err := RunCallbacks(ctx, record, AfterCreate); err != nil {
		return err
	}
	return RunCallbacks(ctx, record, AfterSave)
}

// Update modifies an existing record with callbacks.
func (r *Repository[T]) Update(ctx context.Context, record *T) error {
	if err := RunCallbacks(ctx, record, BeforeUpdate); err != nil {
		return err
	}
	if err := RunCallbacks(ctx, record, BeforeSave); err != nil {
		return err
	}

	columns, _, values := extractFields(record, true)
	setClauses := make([]string, len(columns))
	for i, col := range columns {
		setClauses[i] = fmt.Sprintf("%s = %s", col, database.Placeholder(i+1))
	}

	id := recordID(record)
	values = append(values, id)

	query := fmt.Sprintf(
		"UPDATE %s SET %s, updated_at = %s WHERE id = %s",
		r.TableName,
		joinStrings(setClauses),
		database.NowFunc(),
		database.Placeholder(len(values)),
	)

	if _, err := execContext(ctx, query, values...); err != nil {
		return err
	}

	if err := RunCallbacks(ctx, record, AfterUpdate); err != nil {
		return err
	}
	return RunCallbacks(ctx, record, AfterSave)
}

// Save validates and creates or updates the record.
func (r *Repository[T]) Save(ctx context.Context, record *T) error {
	if err := RunCallbacks(ctx, record, BeforeValidate); err != nil {
		return err
	}

	if v, ok := any(record).(Validatable); ok {
		if errs := Validate(v); errs.Any() {
			return fmt.Errorf("validation failed: %v", errs.Messages())
		}
	}

	if err := RunCallbacks(ctx, record, AfterValidate); err != nil {
		return err
	}

	if recordID(record) == 0 {
		return r.Create(ctx, record)
	}
	return r.Update(ctx, record)
}

// Delete removes a record by ID (soft delete if enabled).
func (r *Repository[T]) Delete(ctx context.Context, id int64) error {
	record, err := r.Find(ctx, id)
	if err != nil {
		return err
	}
	return r.Destroy(ctx, record)
}

// Destroy deletes a record instance with callbacks.
func (r *Repository[T]) Destroy(ctx context.Context, record *T) error {
	if err := RunCallbacks(ctx, record, BeforeDelete); err != nil {
		return err
	}

	if r.softDelete {
		col := r.deletedColumnName()
		_, err := r.Query(ctx).WhereEq("id", recordID(record)).UpdateAll(map[string]any{
			col: database.NowFunc(),
		})
		if err != nil {
			return err
		}
	} else {
		query := fmt.Sprintf("DELETE FROM %s WHERE id = %s", r.TableName, database.Placeholder(1))
		if _, err := execContext(ctx, query, recordID(record)); err != nil {
			return err
		}
	}

	return RunCallbacks(ctx, record, AfterDelete)
}

// Reload refreshes the record from the database.
func (r *Repository[T]) Reload(ctx context.Context, record *T) error {
	fresh, err := r.Find(ctx, recordID(record))
	if err != nil {
		return err
	}
	reflect.ValueOf(record).Elem().Set(reflect.ValueOf(*fresh))
	return nil
}

// CreateRecord creates and saves a record from attributes.
func (r *Repository[T]) CreateRecord(ctx context.Context, attrs map[string]any) (*T, error) {
	record := newRecord[T]()
	applyAttributes(record, attrs)
	if err := r.Save(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

// UpdateRecord updates a record by ID with attributes.
func (r *Repository[T]) UpdateRecord(ctx context.Context, id int64, attrs map[string]any) (*T, error) {
	record, err := r.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	applyAttributes(record, attrs)
	if err := r.Save(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

// FindOrCreateBy finds or creates a record.
func (r *Repository[T]) FindOrCreateBy(ctx context.Context, attrs map[string]any) (*T, bool, error) {
	return r.Query(ctx).FirstOrCreate(attrs)
}

// Touch updates the updated_at timestamp.
func (r *Repository[T]) Touch(ctx context.Context, id int64) error {
	_, err := r.Query(ctx).WhereEq("id", id).Touch()
	return err
}

// Pluck returns values for a single column.
func (r *Repository[T]) Pluck(ctx context.Context, column string) ([]any, error) {
	return r.Query(ctx).Pluck(column)
}

// Exists checks if any record exists.
func (r *Repository[T]) Exists(ctx context.Context) (bool, error) {
	return r.Query(ctx).Exists()
}

// ErrNotFound is returned when a record is not found.
var ErrNotFound = sql.ErrNoRows

package model

import (
	"context"
	"fmt"

	"github.com/lsgser/gofreight/database"
)

// ErrSoftDeleteDisabled is returned when a soft-delete operation is used without EnableSoftDelete.
var ErrSoftDeleteDisabled = fmt.Errorf("model: soft delete is not enabled for this repository")

func (r *Repository[T]) deletedColumnName() string {
	if r.deletedColumn != "" {
		return r.deletedColumn
	}
	return "deleted_at"
}

// SoftDeletesEnabled reports whether soft delete is enabled on this repository.
func (r *Repository[T]) SoftDeletesEnabled() bool {
	return r.softDelete
}

// Restore clears the deleted timestamp for a soft-deleted record.
func (r *Repository[T]) Restore(ctx context.Context, record *T) error {
	return r.RestoreByID(ctx, recordID(record))
}

// RestoreByID restores a soft-deleted record by primary key.
func (r *Repository[T]) RestoreByID(ctx context.Context, id int64) error {
	if !r.softDelete {
		return ErrSoftDeleteDisabled
	}
	col := r.deletedColumnName()
	n, err := r.Query(ctx).WithTrashed().WhereEq("id", id).UpdateAll(map[string]any{col: nil})
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

// ForceDestroy permanently removes a record, bypassing soft delete.
func (r *Repository[T]) ForceDestroy(ctx context.Context, record *T) error {
	if err := RunCallbacks(ctx, record, BeforeDelete); err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = %s", r.TableName, database.Placeholder(1))
	if _, err := execContext(ctx, query, recordID(record)); err != nil {
		return err
	}
	return RunCallbacks(ctx, record, AfterDelete)
}

// ForceDelete permanently removes a record by ID.
func (r *Repository[T]) ForceDelete(ctx context.Context, id int64) error {
	record, err := r.Query(ctx).WithTrashed().Find(id)
	if err != nil {
		return err
	}
	return r.ForceDestroy(ctx, record)
}

// RestoreAll clears deleted_at for all matching rows (must include WithTrashed in the query chain).
func (q *Query[T]) RestoreAll() (int64, error) {
	if !q.withTrashed {
		return 0, fmt.Errorf("model: RestoreAll requires WithTrashed()")
	}
	col := q.deletedColumn
	if col == "" {
		col = "deleted_at"
	}
	return q.UpdateAll(map[string]any{col: nil})
}

// ForceDelete permanently deletes all matching rows.
func (q *Query[T]) ForceDelete() (int64, error) {
	sql := fmt.Sprintf("DELETE FROM %s", q.table)
	args := q.buildWheres(&sql)
	result, err := execContext(q.ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

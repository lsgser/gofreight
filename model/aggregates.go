package model

import (
	"fmt"
	"strings"

	"github.com/lsgser/gofreight/database"
)

// Count returns the number of matching records.
func (q *Query[T]) Count() (int64, error) {
	sql, args := q.buildAggregate("COUNT(*)")
	row := queryRowContext(q.ctx, sql, args...)
	var count int64
	return count, row.Scan(&count)
}

// Sum returns the sum of a column.
func (q *Query[T]) Sum(column string) (float64, error) {
	sql, args := q.buildAggregate(fmt.Sprintf("SUM(%s)", quoteIdent(column)))
	row := queryRowContext(q.ctx, sql, args...)
	var sum float64
	return sum, row.Scan(&sum)
}

// Avg returns the average of a column.
func (q *Query[T]) Avg(column string) (float64, error) {
	sql, args := q.buildAggregate(fmt.Sprintf("AVG(%s)", quoteIdent(column)))
	row := queryRowContext(q.ctx, sql, args...)
	var avg float64
	return avg, row.Scan(&avg)
}

// Min returns the minimum value of a column.
func (q *Query[T]) Min(column string) (any, error) {
	sql, args := q.buildAggregate(fmt.Sprintf("MIN(%s)", quoteIdent(column)))
	row := queryRowContext(q.ctx, sql, args...)
	var min any
	return min, row.Scan(&min)
}

// Max returns the maximum value of a column.
func (q *Query[T]) Max(column string) (any, error) {
	sql, args := q.buildAggregate(fmt.Sprintf("MAX(%s)", quoteIdent(column)))
	row := queryRowContext(q.ctx, sql, args...)
	var max any
	return max, row.Scan(&max)
}

// Pluck returns a slice of values for a single column.
func (q *Query[T]) Pluck(column string) ([]any, error) {
	q.selects = []string{quoteIdent(column)}
	sql, args := q.buildSelect()
	rows, err := queryContext(q.ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []any
	for rows.Next() {
		var val any
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		results = append(results, val)
	}
	return results, rows.Err()
}

// PluckStrings plucks string values from a column.
func (q *Query[T]) PluckStrings(column string) ([]string, error) {
	vals, err := q.Pluck(column)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(vals))
	for i, v := range vals {
		if s, ok := v.(string); ok {
			result[i] = s
		} else if v != nil {
			result[i] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

// PluckInt64s plucks int64 values from a column.
func (q *Query[T]) PluckInt64s(column string) ([]int64, error) {
	vals, err := q.Pluck(column)
	if err != nil {
		return nil, err
	}
	result := make([]int64, len(vals))
	for i, v := range vals {
		switch n := v.(type) {
		case int64:
			result[i] = n
		case int:
			result[i] = int64(n)
		case float64:
			result[i] = int64(n)
		}
	}
	return result, nil
}

func (q *Query[T]) buildAggregate(fn string) (string, []any) {
	sql := fmt.Sprintf("SELECT %s FROM %s", fn, q.table)
	args := q.buildJoins(&sql)
	args = append(args, q.buildWheres(&sql)...)
	return sql, args
}

// Value returns a single scalar value from the first row.
func (q *Query[T]) Value(column string) (any, error) {
	q.limit = 1
	q.selects = []string{quoteIdent(column)}
	sql, args := q.buildSelect()
	row := queryRowContext(q.ctx, sql, args...)
	var val any
	return val, row.Scan(&val)
}

// IDs plucks all primary key values.
func (q *Query[T]) IDs() ([]int64, error) {
	return q.PluckInt64s("id")
}

// UpdateAll updates all matching records without callbacks.
func (q *Query[T]) UpdateAll(attrs map[string]any) (int64, error) {
	if len(attrs) == 0 {
		return 0, nil
	}

	setParts := make([]string, 0, len(attrs)+1)
	args := make([]any, 0, len(attrs)+1)
	i := 1
	for col, val := range attrs {
		setParts = append(setParts, fmt.Sprintf("%s = %s", quoteIdent(col), database.Placeholder(i)))
		args = append(args, val)
		i++
	}
	setParts = append(setParts, fmt.Sprintf("updated_at = %s", database.NowFunc()))

	sql := fmt.Sprintf("UPDATE %s SET %s", q.table, strings.Join(setParts, ", "))
	whereArgs := q.buildWheres(&sql)
	args = append(args, whereArgs...)

	result, err := execContext(q.ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteAll deletes all matching records without loading them.
func (q *Query[T]) DeleteAll() (int64, error) {
	if q.softDelete {
		return q.UpdateAll(map[string]any{"deleted_at": database.NowFunc()})
	}

	sql := fmt.Sprintf("DELETE FROM %s", q.table)
	args := q.buildWheres(&sql)
	result, err := execContext(q.ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Increment increments a numeric column for matching records.
func (q *Query[T]) Increment(column string, amount ...int) (int64, error) {
	n := 1
	if len(amount) > 0 {
		n = amount[0]
	}
	sql := fmt.Sprintf("UPDATE %s SET %s = %s + %d, updated_at = %s", q.table, quoteIdent(column), quoteIdent(column), n, database.NowFunc())
	args := q.buildWheres(&sql)
	result, err := execContext(q.ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Decrement decrements a numeric column.
func (q *Query[T]) Decrement(column string, amount ...int) (int64, error) {
	n := 1
	if len(amount) > 0 {
		n = amount[0]
	}
	return q.Increment(column, -n)
}

// Touch updates the updated_at timestamp for matching records.
func (q *Query[T]) Touch() (int64, error) {
	return q.UpdateAll(map[string]any{})
}

package model

/*
|--------------------------------------------------------------------------
| Query
|--------------------------------------------------------------------------
|
| Implements Query as part of the model package in the Gofreight
| framework. Key symbols: WhereOperator, Query, ScopeFunc, NewQuery,
| Context, Select.
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
| Symbols defined here include: WhereOperator (exported type); OpEq
| (exported value); OpNotEq (exported value); OpGt (exported value); OpGte
| (exported value); OpLt (exported value); OpLte (exported value); OpLike
| (exported value); OpNotLike (exported value); OpIn (exported value);
| OpNotIn (exported value); OpIsNull (exported value); OpIsNotNull
| (exported value); OpBetween (exported value).
| 
*/

import (
	"context"
	"fmt"
	"strings"

	"github.com/lsgser/gofreight/database"
)

// WhereOperator defines SQL comparison operators.
type WhereOperator string

const (
	OpEq        WhereOperator = "="
	OpNotEq     WhereOperator = "!="
	OpGt        WhereOperator = ">"
	OpGte       WhereOperator = ">="
	OpLt        WhereOperator = "<"
	OpLte       WhereOperator = "<="
	OpLike      WhereOperator = "LIKE"
	OpNotLike   WhereOperator = "NOT LIKE"
	OpIn        WhereOperator = "IN"
	OpNotIn     WhereOperator = "NOT IN"
	OpIsNull    WhereOperator = "IS NULL"
	OpIsNotNull WhereOperator = "IS NOT NULL"
	OpBetween   WhereOperator = "BETWEEN"
)

type whereClause struct {
	column   string
	operator WhereOperator
	value    any
	value2   any // for BETWEEN
	orGroup  bool
	raw      string
	rawArgs  []any
}

type orderClause struct {
	column string
	dir    string
}

type joinClause struct {
	joinType string
	table    string
	on       string
}

// Query is a chainable query builder for a repository.
type Query[T any] struct {
	table         string
	ctx           context.Context
	wheres        []whereClause
	orders        []orderClause
	joins         []joinClause
	selects       []string
	groupBy       []string
	having        []whereClause
	limit         int
	offset        int
	preloads      []string
	includes      map[string]bool
	scopes        []string
	withTrashed   bool
	onlyTrashed   bool
	softDelete    bool
	deletedColumn string
	distinct      bool
	lock          string
	repo          *Repository[T]
	scopeFuncs    map[string]ScopeFunc[T]
}

// ScopeFunc is a named scope function.
type ScopeFunc[T any] func(*Query[T]) *Query[T]

// NewQuery creates a query builder for the given table.
func NewQuery[T any](table string, ctx context.Context) *Query[T] {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Query[T]{
		table:         table,
		ctx:           ctx,
		deletedColumn: "deleted_at",
		scopeFuncs:    make(map[string]ScopeFunc[T]),
		includes:      make(map[string]bool),
	}
}

// Context sets the context.
func (q *Query[T]) Context(ctx context.Context) *Query[T] {
	q.ctx = ctx
	return q
}

// Select specifies columns to retrieve.
func (q *Query[T]) Select(columns ...string) *Query[T] {
	q.selects = columns
	return q
}

// Distinct adds DISTINCT to the query.
func (q *Query[T]) Distinct() *Query[T] {
	q.distinct = true
	return q
}

// Where adds a WHERE condition. Supports operators: =, !=, >, >=, <, <=, LIKE, IN, IS NULL, BETWEEN.
func (q *Query[T]) Where(column string, operator WhereOperator, value any) *Query[T] {
	q.wheres = append(q.wheres, whereClause{column: quoteIdent(column), operator: operator, value: value})
	return q
}

// WhereEq is shorthand for Where(column, "=", value).
func (q *Query[T]) WhereEq(column string, value any) *Query[T] {
	return q.Where(column, OpEq, value)
}

// WhereIn adds WHERE column IN (...).
func (q *Query[T]) WhereIn(column string, values []any) *Query[T] {
	return q.Where(column, OpIn, values)
}

// WhereNotIn adds WHERE column NOT IN (...).
func (q *Query[T]) WhereNotIn(column string, values []any) *Query[T] {
	return q.Where(column, OpNotIn, values)
}

// WhereNull adds WHERE column IS NULL.
func (q *Query[T]) WhereNull(column string) *Query[T] {
	return q.Where(column, OpIsNull, nil)
}

// WhereNotNull adds WHERE column IS NOT NULL.
func (q *Query[T]) WhereNotNull(column string) *Query[T] {
	return q.Where(column, OpIsNotNull, nil)
}

// WhereBetween adds WHERE column BETWEEN a AND b.
func (q *Query[T]) WhereBetween(column string, a, b any) *Query[T] {
	q.wheres = append(q.wheres, whereClause{column: quoteIdent(column), operator: OpBetween, value: a, value2: b})
	return q
}

// WhereRaw adds a raw WHERE clause with placeholders.
func (q *Query[T]) WhereRaw(sql string, args ...any) *Query[T] {
	q.wheres = append(q.wheres, whereClause{raw: sql, rawArgs: args})
	return q
}

// OrWhere starts an OR group (simplified: next condition uses OR).
func (q *Query[T]) OrWhere(column string, operator WhereOperator, value any) *Query[T] {
	q.wheres = append(q.wheres, whereClause{column: quoteIdent(column), operator: operator, value: value, orGroup: true})
	return q
}

// Order adds ORDER BY.
func (q *Query[T]) Order(column string, direction ...string) *Query[T] {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	q.orders = append(q.orders, orderClause{column: quoteIdent(column), dir: dir})
	return q
}

// OrderDesc adds ORDER BY column DESC.
func (q *Query[T]) OrderDesc(column string) *Query[T] {
	return q.Order(column, "DESC")
}

// Latest orders by created_at DESC.
func (q *Query[T]) Latest(column ...string) *Query[T] {
	col := "created_at"
	if len(column) > 0 {
		col = column[0]
	}
	return q.OrderDesc(col)
}

// Limit sets LIMIT.
func (q *Query[T]) Limit(n int) *Query[T] {
	q.limit = n
	return q
}

// Offset sets OFFSET.
func (q *Query[T]) Offset(n int) *Query[T] {
	q.offset = n
	return q
}

// Join adds a JOIN clause.
func (q *Query[T]) Join(table, on string) *Query[T] {
	q.joins = append(q.joins, joinClause{joinType: "INNER", table: table, on: on})
	return q
}

// LeftJoin adds a LEFT JOIN.
func (q *Query[T]) LeftJoin(table, on string) *Query[T] {
	q.joins = append(q.joins, joinClause{joinType: "LEFT", table: table, on: on})
	return q
}

// GroupBy adds GROUP BY columns.
func (q *Query[T]) GroupBy(columns ...string) *Query[T] {
	q.groupBy = columns
	return q
}

// Having adds HAVING clause.
func (q *Query[T]) Having(column string, operator WhereOperator, value any) *Query[T] {
	q.having = append(q.having, whereClause{column: quoteIdent(column), operator: operator, value: value})
	return q
}

// With eager-loads associations to prevent N+1 queries.
func (q *Query[T]) With(associations ...string) *Query[T] {
	q.preloads = append(q.preloads, associations...)
	for _, a := range associations {
		q.includes[a] = true
	}
	return q
}

// Scope applies a named scope.
func (q *Query[T]) Scope(name string) *Query[T] {
	q.scopes = append(q.scopes, name)
	if fn, ok := q.scopeFuncs[name]; ok {
		return fn(q)
	}
	if q.repo != nil {
		if fn, ok := q.repo.scopes[name]; ok {
			return fn(q)
		}
	}
	return q
}

// SoftDelete enables soft delete filtering (default: exclude deleted records).
func (q *Query[T]) SoftDelete(column ...string) *Query[T] {
	q.softDelete = true
	if len(column) > 0 {
		q.deletedColumn = column[0]
	}
	return q
}

// WithTrashed includes soft-deleted records.
func (q *Query[T]) WithTrashed() *Query[T] {
	q.withTrashed = true
	return q
}

// OnlyTrashed returns only soft-deleted records.
func (q *Query[T]) OnlyTrashed() *Query[T] {
	q.onlyTrashed = true
	q.withTrashed = true
	return q
}

// ForUpdate adds row locking (SELECT ... FOR UPDATE).
func (q *Query[T]) ForUpdate() *Query[T] {
	q.lock = "FOR UPDATE"
	return q
}

// buildSelect constructs the SELECT SQL and args.
func (q *Query[T]) buildSelect() (string, []any) {
	cols := "*"
	if len(q.selects) > 0 {
		cols = strings.Join(q.selects, ", ")
	}
	if q.distinct {
		cols = "DISTINCT " + cols
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", cols, q.table)
	args := q.buildJoins(&sql)
	args = append(args, q.buildWheres(&sql)...)

	if len(q.groupBy) > 0 {
		sql += " GROUP BY " + strings.Join(q.groupBy, ", ")
	}
	if len(q.having) > 0 {
		havingSQL, havingArgs := buildConditions(q.having, "HAVING")
		sql += " " + havingSQL
		args = append(args, havingArgs...)
	}

	if len(q.orders) > 0 {
		parts := make([]string, len(q.orders))
		for i, o := range q.orders {
			parts[i] = fmt.Sprintf("%s %s", o.column, o.dir)
		}
		sql += " ORDER BY " + strings.Join(parts, ", ")
	} else {
		sql += " ORDER BY id"
	}

	if q.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	if q.offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", q.offset)
	}
	if q.lock != "" {
		sql += " " + q.lock
	}

	return sql, args
}

func (q *Query[T]) buildJoins(sql *string) []any {
	var args []any
	for _, j := range q.joins {
		*sql += fmt.Sprintf(" %s JOIN %s ON %s", j.joinType, j.table, j.on)
	}
	return args
}

func (q *Query[T]) buildWheres(sql *string) []any {
	wheres := q.wheres

	// Soft delete filter
	if q.softDelete && !q.withTrashed {
		wheres = append(wheres, whereClause{column: q.deletedColumn, operator: OpIsNull, value: nil})
	}
	if q.onlyTrashed {
		wheres = append(wheres, whereClause{column: q.deletedColumn, operator: OpIsNotNull, value: nil})
	}

	if len(wheres) == 0 {
		return nil
	}
	whereSQL, args := buildConditions(wheres, "WHERE")
	*sql += " " + whereSQL
	return args
}

func buildConditions(clauses []whereClause, keyword string) (string, []any) {
	if len(clauses) == 0 {
		return "", nil
	}

	var parts []string
	var args []any
	argIndex := 1

	for i, w := range clauses {
		connector := "AND"
		if i > 0 && w.orGroup {
			connector = "OR"
		}

		var part string
		if w.raw != "" {
			part = w.raw
			args = append(args, w.rawArgs...)
		} else {
			switch w.operator {
			case OpIsNull:
				part = fmt.Sprintf("%s IS NULL", w.column)
			case OpIsNotNull:
				part = fmt.Sprintf("%s IS NOT NULL", w.column)
			case OpIn, OpNotIn:
				vals, ok := w.value.([]any)
				if !ok {
					continue
				}
				ph := make([]string, len(vals))
				for j, v := range vals {
					ph[j] = database.Placeholder(argIndex)
					args = append(args, v)
					argIndex++
				}
				part = fmt.Sprintf("%s %s (%s)", w.column, w.operator, strings.Join(ph, ", "))
			case OpBetween:
				part = fmt.Sprintf("%s BETWEEN %s AND %s", w.column, database.Placeholder(argIndex), database.Placeholder(argIndex+1))
				args = append(args, w.value, w.value2)
				argIndex += 2
			default:
				part = fmt.Sprintf("%s %s %s", w.column, w.operator, database.Placeholder(argIndex))
				args = append(args, w.value)
				argIndex++
			}
		}

		if i == 0 {
			parts = append(parts, part)
		} else {
			parts = append(parts, connector+" "+part)
		}
	}

	return keyword + " " + strings.Join(parts, " "), args
}

// Get executes the query and returns all matching records.
func (q *Query[T]) Get() ([]T, error) {
	sql, args := q.buildSelect()
	rows, err := queryContext(q.ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := scanRows[T](rows)
	if err != nil {
		return nil, err
	}

	if len(q.preloads) > 0 && q.repo != nil {
		if err := q.repo.preloadAssociations(q.ctx, results, q.preloads); err != nil {
			return nil, err
		}
	}

	return results, nil
}

// GetCollection executes the query and returns an Eloquent-style collection.
func (q *Query[T]) GetCollection() (Collection[T], error) {
	items, err := q.Get()
	if err != nil {
		return Collection[T]{}, err
	}
	return NewCollection(items), nil
}

// First returns the first matching record or sql.ErrNoRows.
func (q *Query[T]) First() (*T, error) {
	q.limit = 1
	results, err := q.Get()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("record not found")
	}
	return &results[0], nil
}

// Find retrieves a record by primary key.
func (q *Query[T]) Find(id int64) (*T, error) {
	return q.WhereEq("id", id).First()
}

// FindBy retrieves the first record matching a column value.
func (q *Query[T]) FindBy(column string, value any) (*T, error) {
	return q.WhereEq(column, value).First()
}

// Exists returns true if any record matches.
func (q *Query[T]) Exists() (bool, error) {
	count, err := q.Count()
	return count > 0, err
}

// ToSQL returns the generated SQL and args (for debugging).
func (q *Query[T]) ToSQL() (string, []any) {
	return q.buildSelect()
}

package model

import (
	"context"
	"fmt"

	"github.com/lsgser/gofreight/database"
)

// Page holds paginated query results with metadata.
type Page[T any] struct {
	Data        []T   `json:"data"`
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
	From        int   `json:"from"`
	To          int   `json:"to"`
}

// Paginate returns a page of results with metadata.
func (q *Query[T]) Paginate(page, perPage int) (*Page[T], error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	q.limit = perPage
	q.offset = (page - 1) * perPage

	data, err := q.Get()
	if err != nil {
		return nil, err
	}

	lastPage := int(total) / perPage
	if int(total)%perPage > 0 {
		lastPage++
	}
	if lastPage == 0 {
		lastPage = 1
	}

	from := 0
	to := 0
	if len(data) > 0 {
		from = q.offset + 1
		to = q.offset + len(data)
	}

	return &Page[T]{
		Data:        data,
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
		From:        from,
		To:          to,
	}, nil
}

// SimplePaginate returns results without total count (faster for large tables).
func (q *Query[T]) SimplePaginate(page, perPage int) ([]T, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}
	q.limit = perPage + 1 // fetch one extra to detect has-more
	q.offset = (page - 1) * perPage
	return q.Get()
}

// FindEach iterates records in batches to limit memory use.
func (q *Query[T]) FindEach(batchSize int, fn func(T) error) error {
	if batchSize < 1 {
		batchSize = 1000
	}

	offset := 0
	for {
		batch := *q
		batch.limit = batchSize
		batch.offset = offset
		batch.orders = []orderClause{{column: "id", dir: "ASC"}}

		records, err := batch.Get()
		if err != nil {
			return err
		}
		if len(records) == 0 {
			break
		}
		for _, rec := range records {
			if err := fn(rec); err != nil {
				return err
			}
		}
		if len(records) < batchSize {
			break
		}
		offset += batchSize
	}
	return nil
}

// FindOrCreate finds a record by attributes or creates it.
func (q *Query[T]) FindOrCreate(searchAttrs, createAttrs map[string]any) (*T, bool, error) {
	query := q.clone()
	for col, val := range searchAttrs {
		query = query.WhereEq(col, val)
	}

	record, err := query.First()
	if err == nil {
		return record, false, nil
	}

	merged := make(map[string]any)
	for k, v := range searchAttrs {
		merged[k] = v
	}
	for k, v := range createAttrs {
		merged[k] = v
	}

	newRec := newRecord[T]()
	applyAttributes(newRec, merged)

	if err := q.repo.Create(q.ctx, newRec); err != nil {
		return nil, false, err
	}
	return newRec, true, nil
}

// FirstOrCreate returns the first matching record or creates one.
func (q *Query[T]) FirstOrCreate(attrs map[string]any) (*T, bool, error) {
	return q.FindOrCreate(attrs, attrs)
}

// FirstOrInit returns the first match or an initialized (unsaved) record.
func (q *Query[T]) FirstOrInit(attrs map[string]any) (*T, bool, error) {
	query := q.clone()
	for col, val := range attrs {
		query = query.WhereEq(col, val)
	}

	record, err := query.First()
	if err == nil {
		return record, false, nil
	}

	newRec := newRecord[T]()
	applyAttributes(newRec, attrs)
	return newRec, true, nil
}

func (q *Query[T]) clone() *Query[T] {
	c := *q
	c.wheres = append([]whereClause{}, q.wheres...)
	c.orders = append([]orderClause{}, q.orders...)
	c.joins = append([]joinClause{}, q.joins...)
	c.selects = append([]string{}, q.selects...)
	c.preloads = append([]string{}, q.preloads...)
	return &c
}

// Upsert inserts or updates on conflict (PostgreSQL ON CONFLICT / MySQL ON DUPLICATE KEY).
func (r *Repository[T]) Upsert(ctx context.Context, record *T, conflictColumns []string) error {
	columns, placeholders, values := extractFields(record, true)
	colList := joinColumns(columns)

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", r.TableName, colList, joinStrings(placeholders))

	if len(conflictColumns) > 0 && database.DriverName() == "postgres" {
		setParts := make([]string, 0, len(columns))
		for _, col := range columns {
			if col == "id" {
				continue
			}
			setParts = append(setParts, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}
		query += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s", joinColumns(conflictColumns), joinStrings(setParts))
	}

	if database.ReturningClause("id") != "" {
		row := queryRowContext(ctx, query, values...)
		return scanInto(row, record)
	}

	_, err := execContext(ctx, query, values...)
	return err
}

func joinColumns(cols []string) string {
	return joinStrings(cols)
}

func joinStrings(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

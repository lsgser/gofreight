package model

/*
|--------------------------------------------------------------------------
| Transaction
|--------------------------------------------------------------------------
|
| Implements Transaction as part of the model package in the Gofreight
| framework. Key symbols: Transaction, TxFromContext.
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
| Symbols defined here include: Transaction (Transaction runs fn inside a
| database transaction.); TxFromContext (TxFromContext returns the
| transaction from context, or nil.).
| 
*/

import (
	"context"
	"database/sql"

	"github.com/lsgser/gofreight/database"
)

type txKey struct{}

// Transaction runs fn inside a database transaction.
func Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := database.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// TxFromContext returns the transaction from context, or nil.
func TxFromContext(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txKey{}).(*sql.Tx)
	return tx
}

// execContext uses transaction from context if available.
func execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return database.DB().ExecContext(ctx, query, args...)
}

// queryContext uses transaction from context if available.
func queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return database.DB().QueryContext(ctx, query, args...)
}

// queryRowContext uses transaction from context if available.
func queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return database.DB().QueryRowContext(ctx, query, args...)
}

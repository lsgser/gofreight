package model

import (
	"context"
	"database/sql"

	"github.com/gofreight/gofreight/database"
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

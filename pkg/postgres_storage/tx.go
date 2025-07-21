package postgresstorage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transactor определяет интерфейс для работы с транзакциями в PostgreSQL
type Transactor interface {
	WithinTransaction(context.Context, func(ctx context.Context) error) error
}

// Transaction представляет собой менеджер транзакций для работы с PostgreSQL
type Transaction struct {
	db *pgxpool.Pool
}

// NewTransactionManager создает новый менеджер транзакций
func NewTransactionManager(db *pgxpool.Pool) *Transaction {
	return &Transaction{db: db}
}

type txKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func (t *Transaction) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		return err
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("commit failed: %w", commitErr)
	}

	return nil
}

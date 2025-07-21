package postgresstorage

import (
	"context"
	"time"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DBClient представляет клиент для работы с PostgreSQL
type DBClient struct {
	pool   *pgxpool.Pool
	tx     *Transaction
	logger *logger.Logger
}

// NewDBClient создает новый клиент для работы с PostgreSQL
func NewDBClient(pool *pgxpool.Pool, logger *logger.Logger) *DBClient {
	client := &DBClient{
		pool:   pool,
		logger: logger,
	}
	client.tx = NewTransactionManager(pool)
	return client
}

// GetPool возвращает пул соединений
func (c *DBClient) GetPool() *pgxpool.Pool {
	return c.pool
}

// WithinTransaction реализует метод интерфейса Transactor
func (c *DBClient) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	return c.tx.WithinTransaction(ctx, tFunc)
}

// Exec выполняет запрос с логированием
func (c *DBClient) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()

	c.logger.Debug(ctx, "Выполнение SQL запроса",
		zap.String("query", query),
		zap.Any("args", args))

	var result pgconn.CommandTag
	var err error

	tx := extractTx(ctx)
	if tx != nil {
		result, err = tx.Exec(ctx, query, args...)
	} else {
		result, err = c.pool.Exec(ctx, query, args...)
	}

	duration := time.Since(start)
	if err != nil {
		c.logger.Error(ctx, "Ошибка выполнения SQL запроса",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		c.logger.Debug(ctx, "Успешное выполнение SQL запроса",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.String("result", result.String()))
	}

	return result, err
}

// Query выполняет запрос и возвращает результат с логированием
func (c *DBClient) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	start := time.Now()

	c.logger.Debug(ctx, "Выполнение SQL запроса Query",
		zap.String("query", query),
		zap.Any("args", args))

	var rows pgx.Rows
	var err error

	tx := extractTx(ctx)
	if tx != nil {
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = c.pool.Query(ctx, query, args...)
	}

	duration := time.Since(start)
	if err != nil {
		c.logger.Error(ctx, "Ошибка выполнения SQL запроса Query",
			zap.String("query", query),
			zap.Any("args", args),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		c.logger.Debug(ctx, "Успешное выполнение SQL запроса Query",
			zap.String("query", query),
			zap.Duration("duration", duration))
	}

	return rows, err
}

// QueryRow выполняет запрос и возвращает одну строку с логированием
func (c *DBClient) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	start := time.Now()

	c.logger.Debug(ctx, "Выполнение SQL запроса QueryRow",
		zap.String("query", query),
		zap.Any("args", args))

	var row pgx.Row

	tx := extractTx(ctx)
	if tx != nil {
		row = tx.QueryRow(ctx, query, args...)
	} else {
		row = c.pool.QueryRow(ctx, query, args...)
	}

	duration := time.Since(start)
	c.logger.Debug(ctx, "Выполнение SQL запроса QueryRow завершено",
		zap.String("query", query),
		zap.Duration("duration", duration))

	return row
}

// Close закрывает соединения с БД
func (c *DBClient) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

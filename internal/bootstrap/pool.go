package bootstrap

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/config"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
)

// createPool создает пул соединений на основе DSN строки
func createPool(ctx context.Context, dsn string, cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error(ctx, "Ошибка парсинга DSN", zap.Error(err))
		return nil, err
	}

	config.MaxConns = int32(cfg.Postgres.MaxOpenConns)
	config.MinConns = int32(cfg.Postgres.MinConns)
	config.MaxConnLifetime = cfg.Postgres.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Error(ctx, "Ошибка создания пула соединений", zap.Error(err))
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		log.Error(ctx, "Ошибка проверки соединения", zap.Error(err))
		return nil, err
	}

	return pool, nil
}

// NewPool создает новый пул соединений с PostgreSQL
func NewPool(ctx context.Context, cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, error) {
	dsn := cfg.GetPostgresDSN()

	log.Info(ctx, "Подключение к серверу PostgreSQL")

	pool, err := createPool(ctx, dsn, cfg, log)
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "Успешное подключение к PostgreSQL")
	return pool, nil
}

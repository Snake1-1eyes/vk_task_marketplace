package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	app_errors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/entity"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	postgresstorage "github.com/Snake1-1eyes/vk_task_marketplace/pkg/postgres_storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dbManager interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	GetPool() *pgxpool.Pool
}

// Repository реализует интерфейс auth.Repository
type Repository struct {
	db        dbManager
	txManager postgresstorage.Transactor
	logger    *logger.Logger
}

// New создает новый экземпляр репозитория
func New(db dbManager, txManager postgresstorage.Transactor, logger *logger.Logger) *Repository {
	return &Repository{
		db:        db,
		txManager: txManager,
		logger:    logger,
	}
}

// withTransaction выполняет функцию в рамках транзакции, если доступно
func (r *Repository) withTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if r.txManager != nil {
		return r.txManager.WithinTransaction(ctx, fn)
	}
	return fn(ctx)
}

// scanUser сканирует строку базы данных в структуру User
func scanUser(row pgx.Row) (*entity.User, error) {
	user := &entity.User{}
	var createdAt time.Time

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&createdAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("ошибка сканирования пользователя: %w", err)
	}

	user.CreatedAt = createdAt
	return user, nil
}

// CreateUser создает нового пользователя в базе данных
func (r *Repository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := `
		INSERT INTO users (username, password, created_at)
		VALUES ($1, $2, $3)
		RETURNING id, username, password, created_at`

	var result *entity.User
	var err error

	err = r.withTransaction(ctx, func(ctx context.Context) error {
		row := r.db.QueryRow(
			ctx,
			query,
			user.Username,
			user.Password,
			user.CreatedAt,
		)

		result, err = scanUser(row)
		return err
	})

	if err != nil {
		return nil, app_errors.WrapError(err, "ошибка создания пользователя")
	}

	return result, nil
}

// GetUserByUsername находит пользователя по имени
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT id, username, password, created_at
		FROM users
		WHERE username = $1`

	row := r.db.QueryRow(ctx, query, username)
	user, err := scanUser(row)

	if err != nil {
		if errors.Is(err, app_errors.ErrUserNotFound) {
			return nil, app_errors.ErrUserNotFound
		}
		return nil, app_errors.WrapError(err, "ошибка получения пользователя по имени")
	}

	return user, nil
}

// GetUserByID находит пользователя по ID
func (r *Repository) GetUserByID(ctx context.Context, id uint64) (*entity.User, error) {
	query := `
		SELECT id, username, password, created_at
		FROM users
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)
	user, err := scanUser(row)

	if err != nil {
		if errors.Is(err, app_errors.ErrUserNotFound) {
			return nil, app_errors.ErrUserNotFound
		}
		return nil, app_errors.WrapError(err, "ошибка получения пользователя по ID")
	}

	return user, nil
}

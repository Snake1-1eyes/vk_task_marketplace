package postgres

import (
	"context"
	"fmt"
	"strings"

	app_errors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/entity"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/listing"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	postgresstorage "github.com/Snake1-1eyes/vk_task_marketplace/pkg/postgres_storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type dbManager interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	GetPool() *pgxpool.Pool
}

// Repository реализует интерфейс listing.Repository
type Repository struct {
	db        dbManager
	txManager postgresstorage.Transactor
	logger    *logger.Logger
}

// New создает новый экземпляр репозитория
func New(db dbManager, txManager postgresstorage.Transactor, logger *logger.Logger) listing.Repository {
	return &Repository{
		db:        db,
		txManager: txManager,
		logger:    logger,
	}
}

// CreateListing создает новое объявление
func (r *Repository) CreateListing(ctx context.Context, listing *entity.Listing) (*entity.Listing, error) {
	var result *entity.Listing

	err := r.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		insertQuery := `
			INSERT INTO listings (title, description, image_url, price, author_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at`

		var id uint64
		var createdAt pgtype.Timestamptz

		err := r.db.QueryRow(
			txCtx,
			insertQuery,
			listing.Title,
			listing.Description,
			listing.ImageURL,
			listing.Price,
			listing.AuthorID,
			listing.CreatedAt,
		).Scan(&id, &createdAt)

		if err != nil {
			r.logger.Error(ctx, "Ошибка при создании объявления", zap.Error(err))
			return app_errors.WrapError(err, "ошибка при создании объявления")
		}

		listing.ID = id
		listing.CreatedAt = createdAt.Time

		userQuery := `
			SELECT username FROM users WHERE id = $1`

		err = r.db.QueryRow(txCtx, userQuery, listing.AuthorID).Scan(&listing.AuthorUsername)
		if err != nil {
			r.logger.Warn(ctx, "Не удалось получить имя пользователя",
				zap.Uint64("author_id", listing.AuthorID),
				zap.Error(err))
		}

		result = listing
		return nil
	})

	if err != nil {
		return nil, err
	}

	r.logger.Info(ctx, "Объявление успешно создано",
		zap.Uint64("listing_id", result.ID),
		zap.Uint64("author_id", result.AuthorID))

	return result, nil
}

// GetListings получает список объявлений с фильтрацией и пагинацией
func (r *Repository) GetListings(ctx context.Context, filter *entity.ListingFilter) ([]*entity.Listing, uint32, error) {
	baseQuery := `
		FROM listings l
		JOIN users u ON l.author_id = u.id
		WHERE 1=1`

	conditions := []string{}
	args := []any{}
	argIndex := 1

	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("l.price >= $%d", argIndex))
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("l.price <= $%d", argIndex))
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	sortField := "l.created_at"
	if filter.SortBy == "price" {
		sortField = "l.price"
	}

	sortDirection := "DESC"
	if !filter.SortDesc {
		sortDirection = "ASC"
	}

	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause

	dataQuery := `
		SELECT l.id, l.title, l.description, l.image_url, l.price, l.author_id, u.username, l.created_at` + baseQuery + whereClause + `
		ORDER BY ` + sortField + ` ` + sortDirection + `
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, filter.PerPage, (filter.Page-1)*filter.PerPage)

	var total uint32
	err := r.db.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		r.logger.Error(ctx, "Ошибка при подсчете объявлений", zap.Error(err))
		return nil, 0, app_errors.WrapError(err, "ошибка при получении объявлений")
	}

	if total == 0 {
		return []*entity.Listing{}, 0, nil
	}

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error(ctx, "Ошибка при получении списка объявлений", zap.Error(err))
		return nil, 0, app_errors.WrapError(err, "ошибка при получении объявлений")
	}
	defer rows.Close()

	listings := make([]*entity.Listing, 0)

	for rows.Next() {
		listing := &entity.Listing{}
		var createdAt pgtype.Timestamptz

		err := rows.Scan(
			&listing.ID,
			&listing.Title,
			&listing.Description,
			&listing.ImageURL,
			&listing.Price,
			&listing.AuthorID,
			&listing.AuthorUsername,
			&createdAt,
		)

		if err != nil {
			r.logger.Error(ctx, "Ошибка при сканировании строки объявления", zap.Error(err))
			return nil, 0, app_errors.WrapError(err, "ошибка при получении объявлений")
		}

		listing.CreatedAt = createdAt.Time
		listings = append(listings, listing)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error(ctx, "Ошибка при обработке результатов", zap.Error(err))
		return nil, 0, app_errors.WrapError(err, "ошибка при получении объявлений")
	}

	return listings, total, nil
}

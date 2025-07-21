package usecase

import (
	"context"

	app_errors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/entity"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/listing"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"go.uber.org/zap"
)

// UseCase реализует интерфейс listing.UseCase
type UseCase struct {
	repo listing.Repository
	log  *logger.Logger
}

// New создает новый экземпляр UseCase
func New(repo listing.Repository, log *logger.Logger) *UseCase {
	return &UseCase{
		repo: repo,
		log:  log,
	}
}

// CreateListing создает новое объявление
func (uc *UseCase) CreateListing(ctx context.Context, authorID uint64, title, description, imageURL string, price float32) (*entity.Listing, error) {
	listing := entity.NewListing(title, description, imageURL, price, authorID)

	createdListing, err := uc.repo.CreateListing(ctx, listing)
	if err != nil {
		uc.log.Error(ctx, "Ошибка при создании объявления",
			zap.String("title", title),
			zap.Uint64("author_id", authorID),
			zap.Error(err))
		return nil, err
	}

	uc.log.Info(ctx, "Объявление успешно создано",
		zap.Uint64("listing_id", createdListing.ID),
		zap.Uint64("author_id", authorID))

	return createdListing, nil
}

// GetListings получает список объявлений с фильтрацией, сортировкой и пагинацией
func (uc *UseCase) GetListings(ctx context.Context, page, perPage uint32, sortBy string, sortDesc bool, minPrice, maxPrice *float32) ([]*entity.Listing, uint32, error) {
	if minPrice != nil && maxPrice != nil && *minPrice > *maxPrice {
		return nil, 0, app_errors.WrapError(app_errors.ErrValidation, "минимальная цена не может быть больше максимальной")
	}

	validSortFields := map[string]bool{
		"created_at": true,
		"price":      true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	filter := &entity.ListingFilter{
		Page:     page,
		PerPage:  perPage,
		SortBy:   sortBy,
		SortDesc: sortDesc,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
	}

	listings, total, err := uc.repo.GetListings(ctx, filter)
	if err != nil {
		uc.log.Error(ctx, "Ошибка при получении списка объявлений",
			zap.Uint32("page", page),
			zap.Uint32("per_page", perPage),
			zap.Error(err))
		return nil, 0, err
	}

	uc.log.Info(ctx, "Успешно получен список объявлений",
		zap.Uint32("page", page),
		zap.Uint32("per_page", perPage),
		zap.Int("count", len(listings)),
		zap.Uint32("total", total))

	return listings, total, nil
}

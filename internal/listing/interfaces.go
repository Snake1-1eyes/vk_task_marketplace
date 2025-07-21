package listing

import (
	"context"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/entity"
)

type Repository interface {
	CreateListing(ctx context.Context, listing *entity.Listing) (*entity.Listing, error)
	GetListings(ctx context.Context, filter *entity.ListingFilter) ([]*entity.Listing, uint32, error)
}

type UseCase interface {
	CreateListing(ctx context.Context, authorID uint64, title, description, imageURL string, price float32) (*entity.Listing, error)
	GetListings(ctx context.Context, page, perPage uint32, sortBy string, sortDesc bool, minPrice, maxPrice *float32) ([]*entity.Listing, uint32, error)
}

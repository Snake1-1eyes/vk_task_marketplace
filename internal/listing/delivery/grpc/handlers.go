package grpc

import (
	"context"

	"github.com/Snake1-1eyes/vk_task_marketplace/internal/adapter"
	app_errors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/listing"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/middleware"
	listings_pb "github.com/Snake1-1eyes/vk_task_marketplace/pkg/api/listings"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler структура обработчика gRPC запросов
type Handler struct {
	listings_pb.UnimplementedListingsServiceServer
	listingUC listing.UseCase
	log       *logger.Logger
}

// New создает новый экземпляр Handler
func New(listingUC listing.UseCase, log *logger.Logger) *Handler {
	return &Handler{
		listingUC: listingUC,
		log:       log,
	}
}

// CreateListing обрабатывает запрос на создание нового объявления
func (h *Handler) CreateListing(ctx context.Context, req *listings_pb.CreateListingRequest) (*listings_pb.ListingResponse, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		h.log.Warn(ctx, "Попытка создать объявление без авторизации")
		return nil, adapter.MapError(app_errors.ErrUnauthorized)
	}

	listing, err := h.listingUC.CreateListing(
		ctx,
		userID,
		req.Title,
		req.Description,
		req.ImageUrl,
		req.Price,
	)
	if err != nil {
		h.log.Error(ctx, "Ошибка при создании объявления", zap.Error(err))
		return nil, adapter.MapError(err)
	}

	response := &listings_pb.ListingResponse{
		Id:             listing.ID,
		Title:          listing.Title,
		Description:    listing.Description,
		ImageUrl:       listing.ImageURL,
		Price:          listing.Price,
		AuthorUsername: listing.AuthorUsername,
		CreatedAt:      timestamppb.New(listing.CreatedAt),
		IsOwner:        true,
	}

	return response, nil
}

// GetListings обрабатывает запрос на получение списка объявлений
func (h *Handler) GetListings(ctx context.Context, req *listings_pb.GetListingsRequest) (*listings_pb.ListingsResponse, error) {
	userID, _ := middleware.GetUserID(ctx)

	sortBy := "created_at"
	if req.SortBy == listings_pb.SortField_SORT_FIELD_PRICE {
		sortBy = "price"
	}

	sortDesc := true
	if req.SortOrder == listings_pb.SortOrder_SORT_ORDER_ASC {
		sortDesc = false
	}

	listings, total, err := h.listingUC.GetListings(
		ctx,
		req.Page,
		req.PerPage,
		sortBy,
		sortDesc,
		req.MinPrice,
		req.MaxPrice,
	)
	if err != nil {
		h.log.Error(ctx, "Ошибка при получении списка объявлений", zap.Error(err))
		return nil, adapter.MapError(err)
	}

	response := &listings_pb.ListingsResponse{
		Listings:   make([]*listings_pb.ListingResponse, 0, len(listings)),
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: calculateTotalPages(total, req.PerPage),
	}

	for _, listing := range listings {
		isOwner := listing.AuthorID == userID

		listingResponse := &listings_pb.ListingResponse{
			Id:             listing.ID,
			Title:          listing.Title,
			Description:    listing.Description,
			ImageUrl:       listing.ImageURL,
			Price:          listing.Price,
			AuthorUsername: listing.AuthorUsername,
			CreatedAt:      timestamppb.New(listing.CreatedAt),
			IsOwner:        isOwner,
		}

		response.Listings = append(response.Listings, listingResponse)
	}

	return response, nil
}

func calculateTotalPages(total, perPage uint32) uint32 {
	if perPage == 0 {
		return 0
	}

	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}
	return totalPages
}

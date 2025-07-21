package entity

import (
	"time"
)

// Listing представляет модель объявления
type Listing struct {
	ID             uint64    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	ImageURL       string    `json:"image_url"`
	Price          float32   `json:"price"`
	AuthorID       uint64    `json:"author_id"`
	AuthorUsername string    `json:"author_username"`
	CreatedAt      time.Time `json:"created_at"`
}

// ListingFilter представляет фильтр для поиска объявлений
type ListingFilter struct {
	Page     uint32   `json:"page"`
	PerPage  uint32   `json:"per_page"`
	SortBy   string   `json:"sort_by"`
	SortDesc bool     `json:"sort_desc"`
	MinPrice *float32 `json:"min_price,omitempty"`
	MaxPrice *float32 `json:"max_price,omitempty"`
}

// NewListing создает новое объявление
func NewListing(title, description, imageURL string, price float32, authorID uint64) *Listing {
	return &Listing{
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		Price:       price,
		AuthorID:    authorID,
		CreatedAt:   time.Now(),
	}
}

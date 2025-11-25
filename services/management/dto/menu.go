package dto

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// MenuCategoryDto represents a menu category with optional soft delete timestamp.
type MenuCategoryDto struct {
	ID           uuid.UUID  `json:"id"`
	RestaurantID uuid.UUID  `json:"restaurant_id"`
	Name         string     `json:"name"          validate:"required"`
	Description  string     `json:"description"   validate:"required"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

// MenuItemDto represents a menu item with its details and optional uploaded image.
type MenuItemDto struct {
	ID           uuid.UUID             `json:"id"`
	RestaurantID uuid.UUID             `json:"-"            validate:"required"`
	CategoryID   uuid.UUID             `json:"category_id"  validate:"required"`
	Name         string                `json:"name"         validate:"required"`
	Description  string                `json:"description"  validate:"required"`
	PriceInCents int                   `json:"price"        validate:"required,gt=0"`
	IsAvailable  bool                  `json:"is_available"`
	FileHeader   *multipart.FileHeader `json:"-"`
	ImagePath    string                `json:"image_path"`
}

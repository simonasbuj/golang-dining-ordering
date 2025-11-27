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
	RestaurantID uuid.UUID             `json:"-"`
	CategoryID   uuid.UUID             `json:"category_id"    form:"category_id"    validate:"required"`
	Name         string                `json:"name"           form:"name"           validate:"required"`
	Description  string                `json:"description"    form:"description"    validate:"required"`
	PriceInCents int                   `json:"price_in_cents" form:"price_in_cents" validate:"required,gt=0"`
	IsAvailable  bool                  `json:"is_available"   form:"is_available"`
	FileHeader   *multipart.FileHeader `json:"-"              form:"image"`
	ImagePath    string                `json:"image_path"`
}

// CategoryDto represents a menu category containing its items.
type CategoryDto struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Items       []MenuItemDto `json:"items"`
}

// ListMenuItemsDto holds the full list of categories and their items.
type ListMenuItemsDto struct {
	Categories []CategoryDto `json:"categories"`
}

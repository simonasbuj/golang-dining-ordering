package dto

import (
	"mime/multipart"
	"time"
)

// MenuCategoryDto represents a menu category with optional soft delete timestamp.
type MenuCategoryDto struct {
	ID           string     `json:"id"`
	RestaurantID string     `json:"restaurantId"`
	Name         string     `json:"name"         validate:"required"`
	Description  string     `json:"description"  validate:"required"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
}

// MenuItemDto represents a menu item with its details and optional uploaded image.
type MenuItemDto struct {
	ID           string                `json:"id"`
	RestaurantID string                `json:"-"           validate:"required"`
	CategoryID   string                `json:"categoryId"  validate:"required"`
	Name         string                `json:"name"        validate:"required"`
	Description  string                `json:"description" validate:"required"`
	Price        float64               `json:"price"       validate:"required,gt=0"`
	IsAvailable  bool                  `json:"isAvailable"`
	FileHeader   *multipart.FileHeader `json:"-"`
	ImagePath    string                `json:"imagePath"`
}

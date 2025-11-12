package dto

import "time"

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

// Package dto contains data transfer objects for the application.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateRestaurantDto represents the payload for creating a new restaurant.
type CreateRestaurantDto struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"user_id"`
	Name    string    `json:"name"    validate:"required"`
	Address string    `json:"address" validate:"required"`
}

// GetRestaurantsReqDto represents pagination parameters for fetching restaurants.
type GetRestaurantsReqDto struct {
	Page  int32 `query:"page"`
	Limit int32 `query:"limit" validate:"max=100"`
}

// RestaurantItemDto represents a single restaurant in the response.
type RestaurantItemDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}

// GetRestaurantsRespDto represents a paginated list of restaurants.
type GetRestaurantsRespDto struct {
	Page        int32               `json:"page"`
	Limit       int32               `json:"limit"`
	Total       int                 `json:"total"`
	Restaurants []RestaurantItemDto `json:"restaurants"`
}

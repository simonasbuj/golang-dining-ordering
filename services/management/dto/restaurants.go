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

// UpdateRestaurantRequestDto represents the fields for updating a restaurant.
type UpdateRestaurantRequestDto struct {
	ID         uuid.UUID `json:"id"          validate:"required"`
	UserID     uuid.UUID `json:"user_id"     validate:"required"`
	Name       *string   `json:"name"`
	Address    *string   `json:"address"`
	DeleteFlag *bool     `json:"delete_flag"`
}

// UpdateRestaurantResponseDto represents the restaurant data returned after an update.
type UpdateRestaurantResponseDto struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

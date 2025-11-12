// Package dto contains data transfer objects for the application.
package dto

import "time"

// CreateRestaurantDto represents the payload for creating a new restaurant.
type CreateRestaurantDto struct {
	ID      string `json:"id"`
	UserID  string `json:"userId"`
	Name    string `json:"name"    validate:"required"`
	Address string `json:"address" validate:"required"`
}

// GetRestaurantsReqDto represents pagination parameters for fetching restaurants.
type GetRestaurantsReqDto struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// RestaurantItemDto represents a single restaurant in the response.
type RestaurantItemDto struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetRestaurantsRespDto represents a paginated list of restaurants.
type GetRestaurantsRespDto struct {
	Page        int                 `json:"page"`
	Limit       int                 `json:"limit"`
	Total       int                 `json:"total"`
	Restaurants []RestaurantItemDto `json:"restaurants"`
}

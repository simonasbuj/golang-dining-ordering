// Package dto contains data transfer objects for the application.
package dto

// CreateRestaurantDto represents the payload for creating a new restaurant.
type CreateRestaurantDto struct {
	ID      string `json:"id"`
	UserID  string `json:"userId"`
	Name    string `json:"name"    validate:"required"`
	Address string `json:"address" validate:"required"`
}

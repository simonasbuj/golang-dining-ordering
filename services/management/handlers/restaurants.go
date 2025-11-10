// Package handlers contains HTTP handlers for the application.
package handlers

import (
	"github.com/labstack/echo/v4"
)

// RestaurantsHandler handles restaurant-related HTTP requests.
type RestaurantsHandler struct{}

// NewRestaurantsHandler creates a new RestaurantsHandler.
func NewRestaurantsHandler() *RestaurantsHandler {
	return &RestaurantsHandler{}
}

// HandleCreateRestaurant handles creating a new restaurant.
func (h *RestaurantsHandler) HandleCreateRestaurant(_ echo.Context) error {
	return nil
}

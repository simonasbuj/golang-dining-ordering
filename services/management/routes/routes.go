// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	"context"
	handler "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/management/middleware"

	"github.com/labstack/echo/v4"
)

// AddRrestaurantRoutes registers authentication-related HTTP routes.
func AddRrestaurantRoutes(
	_ context.Context,
	e *echo.Echo,
	h *handler.RestaurantsHandler,
	authEndpoint string,
) {
	api := e.Group("/api/v1")

	api.POST("/restaurants", h.HandleCreateRestaurant, middleware.AuthMiddleware(authEndpoint))
	api.GET("/restaurants", h.HandleGetRestaurants)
	api.GET("/restaurants/:id", h.HandleGetRestaurantByID)
}

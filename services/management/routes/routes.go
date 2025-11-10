// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	"context"
	handler "golang-dining-ordering/services/management/handlers"

	"github.com/labstack/echo/v4"
)

// AddRrestaurantRoutes registers authentication-related HTTP routes.
func AddRrestaurantRoutes(_ context.Context, e *echo.Echo, h *handler.RestaurantsHandler) {
	auth := e.Group("/api/v1")

	auth.POST("/restaurants", h.HandleCreateRestaurant)
}

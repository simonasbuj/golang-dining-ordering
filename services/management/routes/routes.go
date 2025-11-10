package routes

import (
	"context"
	"golang-dining-ordering/services/management/handlers"

	"github.com/labstack/echo/v4"
)

// AddRoutes registers authentication-related HTTP routes.
func AddRrestaurantRoutes(_ context.Context, e *echo.Echo, h *handler.RestaurantsHandler) {
	auth := e.Group("/api/v1")

	auth.POST("/restaurants", h.HandleCreateRestaurant)
}

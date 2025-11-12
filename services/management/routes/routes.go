// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	"context"
	handler "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/management/middleware"

	"github.com/labstack/echo/v4"
)

// AddRrestaurantRoutes registers authentication-related HTTP routes.
func AddRrestaurantRoutes(_ context.Context, e *echo.Echo, h *handler.RestaurantsHandler) {
	api := e.Group("/api/v1")

	api.Use(middleware.AuthMiddleware("http://localhost:42069/api/v1/auth/authorize"))

	api.POST("/restaurants", h.HandleCreateRestaurant)
}

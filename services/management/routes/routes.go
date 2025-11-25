// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	authDto "golang-dining-ordering/services/auth/dto"
	handler "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/management/middleware"

	"github.com/labstack/echo/v4"
)

// AddRrestaurantRoutes registers restaurant management related HTTP routes.
func AddRrestaurantRoutes(
	e *echo.Echo,
	h *handler.RestaurantsHandler,
	authEndpoint string,
) {
	api := e.Group("/api/v1")

	api.POST("/restaurants", h.HandleCreateRestaurant,
		middleware.AuthMiddleware(authEndpoint),
		middleware.RoleMiddleware(authDto.RoleManager),
	)
	api.GET("/restaurants", h.HandleGetRestaurants)
	api.GET("/restaurants/:id", h.HandleGetRestaurantByID)
}

// AddMenuRoutes registers restaurant menus management related HTTP routes.
func AddMenuRoutes(
	e *echo.Echo,
	h *handler.MenuHandler,
	authEndpoint string,
) {
	e.Static("/uploads", "uploads")

	api := e.Group("/api/v1/restaurants/:restaurant_id")

	api.POST("/menu/categories", h.HandleAddMenuCategory, middleware.AuthMiddleware(authEndpoint))

	api.POST("/menu/items", h.HandleAddMenuItem, middleware.AuthMiddleware(authEndpoint))
	api.GET("/menu/items", h.HandleGetMenuItems)
}

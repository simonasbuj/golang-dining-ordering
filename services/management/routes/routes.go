// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	authDto "golang-dining-ordering/services/auth/dto"
	handler "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/management/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// AddRestaurantRoutes registers restaurant management related HTTP routes.
func AddRestaurantRoutes(
	e *echo.Echo,
	h *handler.RestaurantsHandler,
	authEndpoint string,
) {
	e.Pre(echoMiddleware.RemoveTrailingSlash())

	api := e.Group("/api/v1/restaurants")

	api.POST("", h.HandleCreateRestaurant,
		middleware.AuthMiddleware(authEndpoint),
		middleware.RoleMiddleware(authDto.RoleManager),
	)
	api.GET("", h.HandleGetRestaurants)
	api.GET("/:id", h.HandleGetRestaurantByID)
	api.PATCH("/:id", h.HandleUpdateRestaurant,
		middleware.AuthMiddleware(authEndpoint),
		middleware.RoleMiddleware(authDto.RoleManager),
	)

	tablesGroup := api.Group("/:id/tables",
		middleware.AuthMiddleware(authEndpoint),
		middleware.RoleMiddleware(authDto.RoleManager),
	)
	tablesGroup.POST("", h.HandleCreateTable)
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

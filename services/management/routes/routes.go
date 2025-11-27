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

	publicAPI := e.Group("/api/v1/restaurants")
	managerAPI := publicAPI.Group("",
		middleware.AuthMiddleware(authEndpoint),
		middleware.RoleMiddleware(authDto.RoleManager),
	)

	managerAPI.POST("", h.HandleCreateRestaurant)
	managerAPI.PATCH("/:id", h.HandleUpdateRestaurant)
	publicAPI.GET("", h.HandleGetRestaurants)
	publicAPI.GET("/:id", h.HandleGetRestaurantByID)

	managerAPI.POST("/:id/tables", h.HandleCreateTable)
}

// AddMenuRoutes registers restaurant menus management related HTTP routes.
func AddMenuRoutes(
	e *echo.Echo,
	h *handler.MenuHandler,
	authEndpoint string,
) {
	e.Static("/uploads", "uploads")

	publicAPI := e.Group("/api/v1/restaurants/:restaurant_id/menu")
	managerAPI := publicAPI.Group("",
		middleware.AuthMiddleware(authEndpoint),
		middleware.RoleMiddleware(authDto.RoleManager),
	)

	managerAPI.POST("categories", h.HandleAddMenuCategory)

	managerAPI.POST("/items", h.HandleAddMenuItem)
	managerAPI.PATCH("/items/:item_id", h.HandleUpdateMenuItem)
	publicAPI.GET("/items", h.HandleGetMenuItems)
}

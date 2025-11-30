// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	// "golang-dining-ordering/pkg/responses"
	// authDto "golang-dining-ordering/services/auth/dto".
	"golang-dining-ordering/services/management/middleware"
	"golang-dining-ordering/services/orders/handler"

	"github.com/labstack/echo/v4"
)

// AddOrdersRoutes registers orders related HTTP routes.
func AddOrdersRoutes(
	e *echo.Echo,
	h *handler.Handler,
	authEndpoint string,
) {
	publicAPI := e.Group("/api/v1/orders")
	// employeeAPI := publicAPI.Group("",
	// 	middleware.AuthMiddleware(authEndpoint),
	// 	middleware.RoleMiddleware(authDto.RoleManager, authDto.RoleWaiter),
	// )

	publicAPI.GET("/current", h.HandleGetCurrentTableOrder)
	publicAPI.GET("/:order_id", h.HandleGetOrder)
	publicAPI.POST("/:order_id/items", h.HandleAddItemToOrder)
	publicAPI.DELETE("/:order_id/items", h.HandleDeleteItemFromOrder)
	publicAPI.PATCH(
		"/:order_id",
		h.HandleUpdateOrder,
		middleware.AuthMiddleware(authEndpoint, false),
	)
}

// Package routes defines HTTP route handlers and route registration for the application.
package routes

import (
	// "golang-dining-ordering/pkg/responses"
	// authDto "golang-dining-ordering/services/auth/dto".
	"golang-dining-ordering/services/management/middleware"
	"golang-dining-ordering/services/orders/handlers"

	"github.com/labstack/echo/v4"
)

// AddOrdersRoutes registers orders related HTTP routes.
func AddOrdersRoutes(
	e *echo.Echo,
	ordersHandler *handlers.OrdersHandler,
	paymentsHandler *handlers.PaymentsHandler,
	websocketHandler *handlers.WebsocketHandler,
	authEndpoint string,
) {
	publicAPI := e.Group("/api/v1/orders")

	publicAPI.GET("/current", ordersHandler.HandleGetCurrentTableOrder)
	publicAPI.GET("/:order_id", ordersHandler.HandleGetOrder)
	publicAPI.POST("/:order_id/items", ordersHandler.HandleAddItemToOrder)
	publicAPI.DELETE("/:order_id/items", ordersHandler.HandleDeleteItemFromOrder)
	publicAPI.PATCH(
		"/:order_id",
		ordersHandler.HandleUpdateOrder,
		middleware.AuthMiddleware(authEndpoint, false),
	)

	publicAPI.POST("/:order_id/payments", paymentsHandler.HandleCreateCheckout)
	publicAPI.POST("/webhooks/payment-success", paymentsHandler.HandleWebhookSuccess)
	publicAPI.GET("/:order_id/ws", websocketHandler.HandleOrderWebsocket)
}

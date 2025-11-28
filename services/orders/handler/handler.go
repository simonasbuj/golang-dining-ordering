// Package handler contains HTTP handler functions for the application.
package handler

import (
	"golang-dining-ordering/pkg/responses"

	"github.com/labstack/echo/v4"
)

// Handler handles orders-related HTTP requests.
type Handler struct{}

// New creates a new Handler for orders.
func New() *Handler {
	return &Handler{}
}

// HandleGetCurrentTableOrder handles getting current order for restaurant table.
func (h *Handler) HandleGetCurrentTableOrder(c echo.Context) error {
	return responses.JSONSuccess(c, "current order my boy from handler", nil)
}

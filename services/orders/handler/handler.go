// Package handler contains HTTP handler functions for the application.
package handler

import (
	"fmt"
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/services/orders/service"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles orders-related HTTP requests.
type Handler struct {
	svc service.Service
}

// New creates a new Handler for orders.
func New(svc service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// HandleGetCurrentTableOrder handles getting current order for restaurant table.
func (h *Handler) HandleGetCurrentTableOrder(c echo.Context) error {
	tableIDString := c.QueryParam("tableId")

	tableID, err := uuid.Parse(tableIDString)
	if err != nil {
		return responses.JSONError(c, "failed to parse tableId from url", err)
	}

	respDto, err := h.svc.GetOrCreateCurrentOrderForTable(c.Request().Context(), tableID)
	if err != nil {
		return responses.JSONError(
			c,
			"failed to get current order for table",
			fmt.Errorf("handling get current table order: %w", err),
			http.StatusInternalServerError,
		)
	}

	return responses.JSONSuccess(c, "current order my boy from handler", respDto)
}

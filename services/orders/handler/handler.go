// Package handler contains HTTP handler functions for the application.
package handler

import (
	"errors"
	"fmt"
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/pkg/validation"
	hndl "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/service"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const orderIDParamName = "order_id"

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

	return responses.JSONSuccess(c, "fetched current order", respDto)
}

// HandleAddItemToOrder handles http request to add item to order.
func (h *Handler) HandleAddItemToOrder(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	var reqDto dto.AddItemToOrderRequestDto

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	respDto, err := h.svc.AddItemToOrder(c.Request().Context(), orderID, reqDto.ItemID)
	if err != nil {
		if errors.Is(err, service.ErrOrderIsNotOpen) ||
			errors.Is(err, service.ErrItemDoesNotBelongToRestaurant) {
			return responses.JSONError(c, err.Error(), err)
		}

		return responses.JSONError(
			c,
			"failed to add item to order",
			err,
			http.StatusInternalServerError,
		)
	}

	return responses.JSONSuccess(c, "item added to order", respDto)
}

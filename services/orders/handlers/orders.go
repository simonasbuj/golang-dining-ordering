// Package handlers contains HTTP handler functions for the application.
package handlers

import (
	"errors"
	"fmt"
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/pkg/validation"
	hndl "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/services"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const orderIDParamName = "order_id"

// OrdersHandler handles orders-related HTTP requests.
type OrdersHandler struct {
	svc services.OrdersService
}

// NewOrdersHandler creates a new Handler for orders.
func NewOrdersHandler(svc services.OrdersService) *OrdersHandler {
	return &OrdersHandler{
		svc: svc,
	}
}

// HandleGetCurrentTableOrder handles getting current order for restaurant table.
func (h *OrdersHandler) HandleGetCurrentTableOrder(c echo.Context) error {
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

// HandleGetOrder handles getting order by id.
func (h *OrdersHandler) HandleGetOrder(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	respDto, err := h.svc.GetOrder(c.Request().Context(), orderID)
	if err != nil {
		return responses.JSONError(c, "failed to fetch order", err, http.StatusInternalServerError)
	}

	return responses.JSONSuccess(c, "fetched order details", respDto)
}

// HandleAddItemToOrder handles http request to add item to order.
func (h *OrdersHandler) HandleAddItemToOrder(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	var reqDto dto.OrderItemRequestDto

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	respDto, err := h.svc.AddItemToOrder(c.Request().Context(), orderID, reqDto.ItemID)
	if err != nil {
		if errors.Is(err, services.ErrOrderIsNotOpen) ||
			errors.Is(err, services.ErrItemDoesNotBelongToRestaurant) {
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

// HandleDeleteItemFromOrder handles http request to delete an item from an order.
func (h *OrdersHandler) HandleDeleteItemFromOrder(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	var reqDto dto.OrderItemRequestDto

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	respDto, err := h.svc.DeleteOrderItem(c.Request().Context(), reqDto.ItemID, orderID)
	if err != nil {
		if errors.Is(err, services.ErrOrderIsNotOpen) {
			return responses.JSONError(c, err.Error(), err)
		}

		return responses.JSONError(
			c,
			"failed to delete item from order",
			err,
			http.StatusInternalServerError,
		)
	}

	return responses.JSONSuccess(c, "deleted item from order", respDto)
}

// HandleUpdateOrder hanldes http request to update an order.
func (h *OrdersHandler) HandleUpdateOrder(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	user, err := hndl.GetUserFromContext(c, false)
	if err != nil {
		return err
	}

	var reqDto dto.UpdateOrderReqDto

	reqDto.OrderID = orderID

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	respDto, err := h.svc.UpdateOrder(c.Request().Context(), &reqDto, user)
	if err != nil {
		if errors.Is(err, services.ErrOrderFinalized) ||
			errors.Is(
				err,
				services.ErrUserCannotEditLockedOrder,
			) || errors.Is(err, services.ErrPayloadEmpty) || errors.Is(err, services.ErrUserCannotEditStatus) {
			return responses.JSONError(c, err.Error(), err)
		}

		return responses.JSONError(c, "failed to update order", err, http.StatusInternalServerError)
	}

	return responses.JSONSuccess(c, "updated order", respDto)
}

// HandleAddWaiter hanldes http request to assign waiter to an order.
func (h *OrdersHandler) HandleAddWaiter(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	user, err := hndl.GetUserFromContext(c)
	if err != nil {
		return err
	}

	err = h.svc.AssignWaiter(c.Request().Context(), orderID, user.UserID)
	if err != nil {
		return responses.JSONError(c, "failed to assign waiter", err)
	}

	return responses.JSONSuccess(c, "waiter assigned to order", nil)
}

// HandleRemoveWaiter hanldes http request to remove waiter from an order.
func (h *OrdersHandler) HandleRemoveWaiter(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	user, err := hndl.GetUserFromContext(c)
	if err != nil {
		return err
	}

	var reqDto dto.RemoveWaiterReqDto

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	err = h.svc.RemoveWaiter(c.Request().Context(), orderID, user.UserID, reqDto.ID)
	if err != nil {
		return responses.JSONError(c, "failed to remove waiter", err)
	}

	return responses.JSONSuccess(c, "waiter removed from order", nil)
}

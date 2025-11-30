package handlers

import (
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/pkg/validation"
	hndl "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/services"
	"net/http"

	"github.com/labstack/echo/v4"
)

// PaymentsHandler handles payments-related HTTP requests.
type PaymentsHandler struct {
	svc services.PaymentsService
}

// NewPaymentsHandler creates a new Handler for orders.
func NewPaymentsHandler(svc services.PaymentsService) *PaymentsHandler {
	return &PaymentsHandler{
		svc: svc,
	}
}

// HandleCreateCheckout handles http request for new checkout session creation for order.
func (h *PaymentsHandler) HandleCreateCheckout(c echo.Context) error {
	orderID, err := hndl.GetUUUIDFromParams(c, orderIDParamName)
	if err != nil {
		return err
	}

	var reqDto dto.CheckoutSessionRequestDto

	err = validation.ValidateDto(c, &reqDto)
	if err != nil {
		return responses.JSONError(c, err.Error(), err)
	}

	respDto, err := h.svc.CreateCheckout(c.Request().Context(), orderID, &reqDto)
	if err != nil {
		return responses.JSONError(
			c,
			"failed to create checkout session",
			err,
			http.StatusInternalServerError,
		)
	}

	return responses.JSONSuccess(c, "going to create new payment session", respDto)
}

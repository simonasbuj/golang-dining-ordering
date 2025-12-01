package handlers

import (
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/pkg/validation"
	hndl "golang-dining-ordering/services/management/handlers"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/services"
	"io"
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

	return responses.JSONSuccess(c, "checkout session created", respDto)
}

// HandleWebhookSuccess handles webhook events for successful payments.
func (h *PaymentsHandler) HandleWebhookSuccess(c echo.Context) error {
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return responses.JSONError(c, "failed to read request payload", err)
	}

	header := c.Request().Header

	respDto, err := h.svc.HandleWebhookSuccess(c.Request().Context(), payload, header)
	if err != nil {
		return responses.JSONError(c, "failed to verify payment", err)
	}

	return responses.JSONSuccess(c, "payment verified and saved", respDto)
}

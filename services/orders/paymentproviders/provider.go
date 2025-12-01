// Package paymentproviders provides implementations of the PaymentProvider
package paymentproviders

import (
	"context"
	"golang-dining-ordering/services/orders/dto"
	"net/http"
)

// PaymentProvider is an interface for payment-related operations.
type PaymentProvider interface {
	CreateCheckoutSession(
		ctx context.Context,
		reqDto *dto.CheckoutSessionRequestDto,
	) (*dto.CheckoutSessionResponseDto, error)
	VerifySuccessWebhookEvent(
		payload []byte,
		header http.Header,
	) (*dto.PaymentSuccessWebhookResponseDto, error)
}

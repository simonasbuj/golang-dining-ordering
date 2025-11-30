// Package paymentproviders provides implementations of the PaymentProvider
package paymentproviders

import (
	"context"
	"golang-dining-ordering/services/orders/dto"
)

// PaymentProvider is an interface for payment-related operations.
type PaymentProvider interface {
	CreateCheckoutSession(
		ctx context.Context,
		reqDto *dto.CheckoutSessionRequestDto,
	) (string, error)
}

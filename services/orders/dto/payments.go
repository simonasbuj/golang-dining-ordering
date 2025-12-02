package dto

import (
	db "golang-dining-ordering/services/orders/db/generated"

	"github.com/google/uuid"
)

// CheckoutSessionRequestDto represents the data needed to create a checkout session.
type CheckoutSessionRequestDto struct {
	OrderDto   *OrderDto `json:"order"`
	SuccessURL string    `json:"success_url" validate:"required"`
	CancelURL  string    `json:"cancel_url"  validate:"required"`
}

// CheckoutSessionResponseDto represents the response returned after creating a checkout session.
type CheckoutSessionResponseDto struct {
	URL      string                   `json:"url"`
	Provider db.OrdersPaymentProvider `json:"provider"`
}

// PaymentDto represents save payment request and response.
type PaymentDto struct {
	ID                uuid.UUID                `json:"id"`
	OrderID           uuid.UUID                `json:"order_id"`
	AmountInCents     int                      `json:"amount_in_cents"`
	Provider          db.OrdersPaymentProvider `json:"provider"`
	ProviderPaymentID string                   `json:"provider_payment_id"`
	Currency          string                   `json:"currency"`
}

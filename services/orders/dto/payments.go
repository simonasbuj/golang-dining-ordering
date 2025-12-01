package dto

import "github.com/google/uuid"

// CheckoutSessionRequestDto represents the data needed to create a checkout session.
type CheckoutSessionRequestDto struct {
	OrderDto   *OrderDto `json:"order"`
	SuccessURL string    `json:"success_url" validate:"required"`
	CancelURL  string    `json:"cancel_url"  validate:"required"`
}

// CheckoutSessionResponseDto represents the response returned after creating a checkout session.
type CheckoutSessionResponseDto struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

// PaymentSuccessWebhookResponseDto represents the response returned after successful payment webhook is handled.
type PaymentSuccessWebhookResponseDto struct {
	OrderID uuid.UUID `json:"order_id"`
}

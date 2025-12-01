package dto

// CheckoutSessionRequestDto represents the data needed to create a checkout session.
type CheckoutSessionRequestDto struct {
	OrderDto   *OrderDto `json:"order"`
	SuccessURL string    `json:"success_url" validate:"required"`
	CancelURL  string    `json:"cancel_url"  validate:"required"`
}

// CheckoutSessionResponseDto represents the response returned after creating a checkout session.
type CheckoutSessionResponseDto struct {
	URL string `json:"url"`
}

// PaymentSuccessWebhookResponseDto represents the response returned after successful payment webhook is handled.
type PaymentSuccessWebhookResponseDto struct {
	OrderID string `json:"order_id"`
}

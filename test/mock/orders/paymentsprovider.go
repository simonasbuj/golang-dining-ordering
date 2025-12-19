package orders

import (
	"context"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"net/http"
)

type mockPaymentsProvider struct {
	provider db.OrdersPaymentProvider
}

func NewMockPaymentsProvider() *mockPaymentsProvider { //nolint:revive
	return &mockPaymentsProvider{
		provider: testPaymentProvider,
	}
}

func (p *mockPaymentsProvider) CreateCheckoutSession(
	_ context.Context,
	req *dto.CheckoutSessionRequestDto,
) (*dto.CheckoutSessionResponseDto, error) {
	if req.SuccessURL == "" {
		return nil, ErrPaymentProviderFailed
	}

	return &dto.CheckoutSessionResponseDto{
		URL:      testCheckoutURL,
		Provider: p.provider,
	}, nil
}

func (p *mockPaymentsProvider) VerifySuccessWebhookEvent(
	payload []byte,
	_ http.Header,
) (*dto.PaymentDto, error) {
	if len(payload) == 0 {
		return nil, ErrPaymentProviderFailed
	}

	if len(payload) == 1 {
		return &dto.PaymentDto{}, nil
	}

	return &dto.PaymentDto{
		ID:                testPaymentID,
		OrderID:           testOrderID,
		AmountInCents:     testAmount,
		Provider:          p.provider,
		ProviderPaymentID: testProviderPaymentID,
		Currency:          testCurrency,
	}, nil
}

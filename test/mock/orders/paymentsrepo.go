package orders

import (
	"context"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"

	"github.com/google/uuid"
)

type mockPaymentsRepo struct{}

func NewMockPaymentsRepo() *mockPaymentsRepo { //nolint:revive
	return &mockPaymentsRepo{}
}

func (r *mockPaymentsRepo) SavePayment(
	_ context.Context,
	reqDto *dto.PaymentDto,
) (*dto.PaymentDto, error) {
	if reqDto.OrderID == uuid.Nil {
		return nil, ErrRepoFailed
	}

	return &dto.PaymentDto{
		ID:                testPaymentID,
		OrderID:           testOrderID,
		AmountInCents:     testAmount,
		Provider:          db.OrdersPaymentProviderMock,
		ProviderPaymentID: testProviderPaymentID,
		Currency:          testCurrency,
	}, nil
}

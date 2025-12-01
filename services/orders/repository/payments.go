package repository

import (
	"context"
	"fmt"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"

	"github.com/google/uuid"
)

// PaymentsRepo defines methods for accessing and managing payments data.
type PaymentsRepo interface {
	SavePayment(ctx context.Context, reqDto *dto.PaymentDto) (*dto.PaymentDto, error)
}

type paymentsRepo struct {
	q *db.Queries
}

// NewPaymentsRepo creates a new payments reposiotry instance.
//
//revive:disable:unexported-return
func NewPaymentsRepo(q *db.Queries) *paymentsRepo {
	return &paymentsRepo{
		q: q,
	}
}

//revive:enable:unexported-return

func (r *paymentsRepo) SavePayment(
	ctx context.Context,
	reqDto *dto.PaymentDto,
) (*dto.PaymentDto, error) {
	row, err := r.q.SavePayment(ctx, db.SavePaymentParams{
		ID:                uuid.New(),
		OrderID:           reqDto.OrderID,
		AmountInCents:     reqDto.AmountInCents,
		Currency:          reqDto.Currency,
		Provider:          reqDto.Provider,
		ProviderPaymentID: reqDto.ProviderPaymentID,
	})
	if err != nil {
		return nil, fmt.Errorf("saving payment %+v to database: %w", reqDto, err)
	}

	return &dto.PaymentDto{
		ID:                row.ID,
		OrderID:           row.OrderID,
		AmountInCents:     row.AmountInCents,
		Currency:          row.Currency,
		Provider:          row.Provider,
		ProviderPaymentID: row.ProviderPaymentID,
	}, nil
}

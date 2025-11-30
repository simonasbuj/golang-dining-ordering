package services

import (
	"context"
	"errors"
	"fmt"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/paymentproviders"
	"golang-dining-ordering/services/orders/repository"

	"github.com/google/uuid"
)

// PaymentsService defines business logic methods for payments service.
type PaymentsService interface {
	CreateCheckout(
		ctx context.Context,
		orderID uuid.UUID,
		reqDto *dto.CheckoutSessionRequestDto,
	) (*dto.CheckoutSessionResponseDto, error)
}

// ErrOrderPriceIsZero is returned when order's total amount and tip are 0.
var ErrOrderPriceIsZero = errors.New("order total price and tip amount are 0")

type paymentsService struct {
	ordersRepo repository.OrdersRepo
	provider   paymentproviders.PaymentProvider
}

// NewPaymentsService creates a new payments service instance.
//
//revive:disable:unexported-return
func NewPaymentsService(
	ordersRepo repository.OrdersRepo,
	provider paymentproviders.PaymentProvider,
) *paymentsService {
	return &paymentsService{
		ordersRepo: ordersRepo,
		provider:   provider,
	}
}

//revive:enable:unexported-return

func (s *paymentsService) CreateCheckout(
	ctx context.Context,
	orderID uuid.UUID,
	reqDto *dto.CheckoutSessionRequestDto,
) (*dto.CheckoutSessionResponseDto, error) {
	order, err := s.ordersRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("getting order: %w", err)
	}

	canPay, err := s.canPayForOrder(order)
	if !canPay || err != nil {
		return nil, err
	}

	reqDto.OrderDto = order

	checkoutURL, err := s.provider.CreateCheckoutSession(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("creating checkout session: %w", err)
	}

	respDto := &dto.CheckoutSessionResponseDto{URL: checkoutURL}

	return respDto, nil
}

func (s *paymentsService) canPayForOrder(order *dto.OrderDto) (bool, error) {
	if order.Status == db.OrderStatusCancelled || order.Status == db.OrderStatusCompleted {
		return false, ErrOrderFinalized
	}

	if order.TipAmountInCents == 0 && order.TotalPriceInCents == 0 {
		return false, ErrOrderPriceIsZero
	}

	return true, nil
}

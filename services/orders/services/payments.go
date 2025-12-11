package services

import (
	"context"
	"errors"
	"fmt"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/paymentproviders"
	"golang-dining-ordering/services/orders/repository"
	"net/http"

	"github.com/google/uuid"
)

// PaymentsService defines business logic methods for payments service.
type PaymentsService interface {
	CreateCheckout(
		ctx context.Context,
		orderID uuid.UUID,
		reqDto *dto.CheckoutSessionRequestDto,
	) (*dto.CheckoutSessionResponseDto, error)
	HandleWebhookSuccess(
		ctx context.Context,
		payload []byte,
		header http.Header,
	) (*dto.PaymentDto, error)
}

// ErrOrderPriceIsZero is returned when order's total amount and tip are 0.
var ErrOrderPriceIsZero = errors.New("order total price and tip amount are 0")

type paymentsService struct {
	ordersRepo   repository.OrdersRepo
	paymentsRepo repository.PaymentsRepo
	provider     paymentproviders.PaymentProvider
}

// NewPaymentsService creates a new payments service instance.
//
//revive:disable:unexported-return
func NewPaymentsService(
	ordersRepo repository.OrdersRepo,
	paymentsRepo repository.PaymentsRepo,
	provider paymentproviders.PaymentProvider,
) *paymentsService {
	return &paymentsService{
		ordersRepo:   ordersRepo,
		paymentsRepo: paymentsRepo,
		provider:     provider,
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

	respDto, err := s.provider.CreateCheckoutSession(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("creating checkout session: %w", err)
	}

	return respDto, nil
}

func (s *paymentsService) HandleWebhookSuccess(
	ctx context.Context,
	payload []byte,
	header http.Header,
) (*dto.PaymentDto, error) {
	paymentDto, err := s.provider.VerifySuccessWebhookEvent(payload, header)
	if err != nil {
		return nil, fmt.Errorf("verifying payment success webhook event: %w", err)
	}

	respDto, err := s.paymentsRepo.SavePayment(ctx, paymentDto)
	if err != nil {
		return nil, fmt.Errorf("creating payment: %w", err)
	}

	status := db.OrderStatusCompleted

	_, err = s.ordersRepo.UpdateOrder(ctx, &dto.UpdateOrderReqDto{
		OrderID:          respDto.OrderID,
		Status:           &status,
		TipAmountInCents: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("updating order status: %w", err)
	}

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

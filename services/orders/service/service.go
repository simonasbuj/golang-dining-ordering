// Package service provides business logic for orders creation and management.
package service

import (
	"context"
	"errors"
	"fmt"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/repository"

	"github.com/google/uuid"
)

var (
	// ErrItemDoesNotBelongToRestaurant is returned when the item is from a different restaurant.
	ErrItemDoesNotBelongToRestaurant = errors.New("item does not belong to this restaurant")
	// ErrOrderIsNotOpen returned when an operation is attempted on a finished or locked order.
	ErrOrderIsNotOpen = errors.New("order is not open")
)

// Service defines business logic methods for orders service.
type Service interface {
	GetOrCreateCurrentOrderForTable(
		ctx context.Context,
		tableID uuid.UUID,
	) (*dto.CurrentOrderDto, error)
	AddItemToOrder(ctx context.Context, orderID, itemID uuid.UUID) (*dto.OrderDto, error)
	DeleteOrderItem(ctx context.Context, orderItemID, orderID uuid.UUID) (*dto.OrderDto, error)
}

type service struct {
	repo repository.Repo
}

// New creates a new orders service instance.
//
//revive:disable:unexported-return
func New(repo repository.Repo) *service {
	return &service{
		repo: repo,
	}
}

//revive:enable:unexported-return

func (s *service) GetOrCreateCurrentOrderForTable(
	ctx context.Context,
	tableID uuid.UUID,
) (*dto.CurrentOrderDto, error) {
	respDto, err := s.repo.GetCurrentOrderForTable(ctx, tableID)
	if err == nil {
		return respDto, nil
	}

	if !errors.Is(err, repository.ErrNoCurrentOrder) {
		return nil, fmt.Errorf("getting current order for table: %w", err)
	}

	currency, err := s.repo.GetTableCurrency(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("getting table currency: %w", err)
	}

	respDto, err = s.repo.CreateOrderForTable(ctx, tableID, currency)
	if err != nil {
		return nil, fmt.Errorf("creating new order: %w", err)
	}

	return respDto, nil
}

func (s *service) AddItemToOrder(
	ctx context.Context,
	orderID, itemID uuid.UUID,
) (*dto.OrderDto, error) {
	item, err := s.repo.GetMenuItem(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("getting menu item: %w", err)
	}

	currentOrder, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("getting current order: %w", err)
	}

	if item.RestaurantID != currentOrder.RestaurantID {
		return nil, ErrItemDoesNotBelongToRestaurant
	}

	if currentOrder.Status != string(db.OrderStatusOpen) {
		return nil, ErrOrderIsNotOpen
	}

	_, err = s.repo.AddItemToOrder(ctx, orderID, item)
	if err != nil {
		return nil, fmt.Errorf("adding item to order: %w", err)
	}

	respDto, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("getting updated order: %w", err)
	}

	return respDto, nil
}

func (s *service) DeleteOrderItem(
	ctx context.Context,
	orderItemID, orderID uuid.UUID,
) (*dto.OrderDto, error) {
	err := s.repo.DeleteOrderItem(ctx, orderItemID, orderID)
	if err != nil {
		return nil, fmt.Errorf("deleting order item: %w", err)
	}

	respDto, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("getting updated order items: %w", err)
	}

	return respDto, nil
}

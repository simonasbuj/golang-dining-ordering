// Package services provides business logic for orders and payments creation and management.
package services

import (
	"context"
	"errors"
	"fmt"
	authDto "golang-dining-ordering/services/auth/dto"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/repository"

	"github.com/google/uuid"
)

// OrdersService defines business logic methods for orders service.
type OrdersService interface {
	GetOrCreateCurrentOrderForTable(
		ctx context.Context,
		tableID uuid.UUID,
	) (*dto.CurrentOrderDto, error)
	GetOrder(ctx context.Context, orderID uuid.UUID) (*dto.OrderDto, error)
	AddItemToOrder(ctx context.Context, orderID, itemID uuid.UUID) (*dto.OrderDto, error)
	DeleteOrderItem(ctx context.Context, orderItemID, orderID uuid.UUID) (*dto.OrderDto, error)
	UpdateOrder(
		ctx context.Context,
		reqDto *dto.UpdateOrderReqDto,
		claims *authDto.TokenClaimsDto,
	) (*dto.OrderDto, error)
	AssignWaiter(ctx context.Context, orderID, userID uuid.UUID) error
	RemoveWaiter(ctx context.Context, orderID, userID, assignID uuid.UUID) error
}

var (
	// ErrItemDoesNotBelongToRestaurant is returned when the item is from a different restaurant.
	ErrItemDoesNotBelongToRestaurant = errors.New("item does not belong to this restaurant")
	// ErrOrderIsNotOpen is returned when an operation is attempted on a finished or locked order.
	ErrOrderIsNotOpen = errors.New("order is not open")
	// ErrPayloadEmpty is returned when all fields in payload are empty.
	ErrPayloadEmpty = errors.New("payload is empty")
	// ErrOrderFinalized is returned when an order cannot be modified because its is completed or canceled.
	ErrOrderFinalized = errors.New("order cannot be edited anymore since it's finalized")
	// ErrUserCannotEditLockedOrder is returned when the current user is not allowed to edit an order that is locked.
	ErrUserCannotEditLockedOrder = errors.New("this user cannot edit locked orders")
	// ErrUserCannotEditStatus is returned when the current user is not allowed to edit status of the order.
	ErrUserCannotEditStatus = errors.New("user cannot edit status of this order")
)

type ordersService struct {
	repo repository.OrdersRepo
}

// NewOrdersService creates a new orders service instance.
//
//revive:disable:unexported-return
func NewOrdersService(repo repository.OrdersRepo) *ordersService {
	return &ordersService{
		repo: repo,
	}
}

//revive:enable:unexported-return

func (s *ordersService) GetOrCreateCurrentOrderForTable(
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

func (s *ordersService) GetOrder(ctx context.Context, orderID uuid.UUID) (*dto.OrderDto, error) {
	respDto, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("getting order: %w", err)
	}

	return respDto, nil
}

func (s *ordersService) AddItemToOrder(
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

	if currentOrder.Status != db.OrderStatusOpen {
		return nil, ErrOrderIsNotOpen
	}

	addedOrderItem, err := s.repo.AddItemToOrder(ctx, orderID, item)
	if err != nil {
		return nil, fmt.Errorf("adding item to order: %w", err)
	}

	currentOrder.Items = append(currentOrder.Items, addedOrderItem)
	currentOrder.TotalPriceInCents += addedOrderItem.PriceInCents

	return currentOrder, nil
}

func (s *ordersService) DeleteOrderItem(
	ctx context.Context,
	orderItemID, orderID uuid.UUID,
) (*dto.OrderDto, error) {
	currentOrder, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("getting current order: %w", err)
	}

	if currentOrder.Status != db.OrderStatusOpen {
		return nil, ErrOrderIsNotOpen
	}

	deletedItem, err := s.repo.DeleteOrderItem(ctx, orderItemID, orderID)
	if err != nil {
		return nil, fmt.Errorf("deleting order item: %w", err)
	}

	for i, item := range currentOrder.Items {
		if item.ID == deletedItem.ID {
			currentOrder.Items = append(
				currentOrder.Items[:i],
				currentOrder.Items[i+1:]...,
			)

			break
		}
	}

	currentOrder.TotalPriceInCents -= deletedItem.PriceInCents

	return currentOrder, nil
}

func (s *ordersService) UpdateOrder(
	ctx context.Context,
	reqDto *dto.UpdateOrderReqDto,
	claims *authDto.TokenClaimsDto,
) (*dto.OrderDto, error) {
	if reqDto.Status == nil && reqDto.TipAmountInCents == nil {
		return nil, ErrPayloadEmpty
	}

	currentOrder, err := s.repo.GetOrderItems(ctx, reqDto.OrderID)
	if err != nil {
		return nil, fmt.Errorf("getting current order: %w", err)
	}

	canEdit, err := s.canUserEditOrder(ctx, currentOrder, claims, reqDto)
	if !canEdit || err != nil {
		return nil, err
	}

	respDto, err := s.repo.UpdateOrder(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("updating order: %w", err)
	}

	currentOrder.Status = respDto.Status
	currentOrder.TipAmountInCents = respDto.TipAmountInCents

	return currentOrder, nil
}

func (s *ordersService) AssignWaiter(ctx context.Context, orderID, userID uuid.UUID) error {
	order, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return fmt.Errorf("fetching order details: %w", err)
	}

	err = s.repo.IsUserRestaurantWaiter(ctx, userID, order.RestaurantID)
	if err != nil {
		return fmt.Errorf("checking if user is restaurant waiter: %w", err)
	}

	err = s.repo.AssignWaiter(ctx, orderID, userID)
	if err != nil {
		return fmt.Errorf("assigning waiter to order: %w", err)
	}

	return nil
}

func (s *ordersService) RemoveWaiter(
	ctx context.Context,
	orderID, userID, assignID uuid.UUID,
) error {
	order, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return fmt.Errorf("fetching order details: %w", err)
	}

	err = s.repo.IsUserRestaurantWaiter(ctx, userID, order.RestaurantID)
	if err != nil {
		return fmt.Errorf("checking if user is restaurant waiter: %w", err)
	}

	err = s.repo.RemoveWaiter(ctx, orderID, userID, assignID)
	if err != nil {
		return fmt.Errorf("removing waiter from order: %w", err)
	}

	return nil
}

func (s *ordersService) canUserEditOrder(
	ctx context.Context,
	order *dto.OrderDto,
	claims *authDto.TokenClaimsDto,
	reqDto *dto.UpdateOrderReqDto,
) (bool, error) {
	if s.isOrderFinalized(order) {
		return false, ErrOrderFinalized
	}

	if !s.canUserChangeStatus(reqDto, claims) {
		return false, ErrUserCannotEditStatus
	}

	if !s.canUserEditLockedOrder(ctx, order, claims, reqDto) {
		return false, ErrUserCannotEditLockedOrder
	}

	return true, nil
}

func (s *ordersService) isOrderFinalized(order *dto.OrderDto) bool {
	return order.Status == db.OrderStatusCancelled || order.Status == db.OrderStatusCompleted
}

func (s *ordersService) canUserEditLockedOrder(
	ctx context.Context,
	order *dto.OrderDto,
	claims *authDto.TokenClaimsDto,
	reqDto *dto.UpdateOrderReqDto,
) bool {
	if order.Status == db.OrderStatusLocked && reqDto.Status != nil {
		if claims.UserID == uuid.Nil {
			return false
		}

		err := s.repo.IsUserRestaurantWaiter(ctx, claims.UserID, order.RestaurantID)
		if err != nil {
			return false
		}
	}

	return true
}

func (s *ordersService) canUserChangeStatus(
	reqDto *dto.UpdateOrderReqDto,
	claims *authDto.TokenClaimsDto,
) bool {
	if reqDto.Status != nil && *reqDto.Status != db.OrderStatusLocked && claims.UserID == uuid.Nil {
		return false
	}

	return true
}

// Package repository provides methods to access and manage orders data from database.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"

	"github.com/google/uuid"
)

var (
	// ErrNoCurrentOrder is returned when there is no active order for the specified table.
	ErrNoCurrentOrder = errors.New("current order for this table doesnt exist")
	// ErrOrderDoesNotExist is returned if order doesn't exist in database.
	ErrOrderDoesNotExist = errors.New("order with this id does not exist")
)

// OrdersRepo defines methods for accessing and managing orders data.
type OrdersRepo interface {
	GetCurrentOrderForTable(ctx context.Context, tableID uuid.UUID) (*dto.CurrentOrderDto, error)
	CreateOrderForTable(
		ctx context.Context,
		tableID uuid.UUID,
		currency string,
	) (*dto.CurrentOrderDto, error)
	GetTableCurrency(ctx context.Context, tableID uuid.UUID) (string, error)
	AddItemToOrder(
		ctx context.Context,
		orderID uuid.UUID,
		item *dto.OrderItemDto,
	) (*dto.OrderItemDto, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) (*dto.OrderDto, error)
	GetMenuItem(ctx context.Context, itemID uuid.UUID) (*dto.OrderItemDto, error)
	DeleteOrderItem(ctx context.Context, orderItemID, orderID uuid.UUID) (*dto.OrderItemDto, error)
	UpdateOrder(ctx context.Context, reqDto *dto.UpdateOrderReqDto) (*dto.OrderDto, error)
	IsUserRestaurantWaiter(ctx context.Context, userID, restaurantID uuid.UUID) error
	AssignWaiter(ctx context.Context, orderID, userID uuid.UUID) error
	RemoveWaiter(ctx context.Context, orderID, userID, assignID uuid.UUID) error
}

type ordersRepo struct {
	q *db.Queries
}

// NewOrdersRepo creates a new orders reposiotry instance.
//
//revive:disable:unexported-return
func NewOrdersRepo(q *db.Queries) *ordersRepo {
	return &ordersRepo{
		q: q,
	}
}

//revive:enable:unexported-return

func (r *ordersRepo) GetCurrentOrderForTable(
	ctx context.Context,
	tableID uuid.UUID,
) (*dto.CurrentOrderDto, error) {
	id, err := r.q.GetCurrentOrder(ctx, tableID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoCurrentOrder
		}

		return nil, fmt.Errorf("fetching current order for table from database: %w", err)
	}

	return &dto.CurrentOrderDto{ID: id}, nil
}

func (r *ordersRepo) CreateOrderForTable(
	ctx context.Context,
	tableID uuid.UUID,
	currency string,
) (*dto.CurrentOrderDto, error) {
	id, err := r.q.CreateOrder(ctx, db.CreateOrderParams{
		ID:       uuid.New(),
		TableID:  tableID,
		Currency: currency,
	})
	if err != nil {
		return nil, fmt.Errorf("inserting new order to database: %w", err)
	}

	return &dto.CurrentOrderDto{ID: id}, nil
}

func (r *ordersRepo) GetTableCurrency(ctx context.Context, tableID uuid.UUID) (string, error) {
	currency, err := r.q.GetTableCurrency(ctx, tableID)
	if err != nil {
		return "", fmt.Errorf("fetching table currency from database: %w", err)
	}

	return currency, nil
}

func (r *ordersRepo) AddItemToOrder(
	ctx context.Context,
	orderID uuid.UUID,
	item *dto.OrderItemDto,
) (*dto.OrderItemDto, error) {
	row, err := r.q.AddOrderItem(ctx, db.AddOrderItemParams{
		ID:           uuid.New(),
		OrderID:      orderID,
		ItemID:       uuid.NullUUID{UUID: item.ID, Valid: true},
		ItemName:     item.Name,
		PriceInCents: item.PriceInCents,
	})
	if err != nil {
		return nil, fmt.Errorf("inserting order item into database: %w", err)
	}

	respDto := &dto.OrderItemDto{
		ID:           row.ID,
		RestaurantID: item.RestaurantID,
		ItemID:       row.ItemID.UUID,
		Name:         row.ItemName,
		PriceInCents: row.PriceInCents,
	}

	return respDto, nil
}

func (r *ordersRepo) GetOrderItems(ctx context.Context, orderID uuid.UUID) (*dto.OrderDto, error) {
	rows, err := r.q.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("getting order items from database: %w", err)
	}

	if len(rows) == 0 {
		return nil, ErrOrderDoesNotExist
	}

	firstRow := rows[0]

	respDto := &dto.OrderDto{
		ID:                firstRow.ID,
		RestaurantID:      firstRow.RestaurantID.UUID,
		RestaurantName:    firstRow.RestaurantName.String,
		Status:            firstRow.Status,
		Currency:          firstRow.Currency,
		TipAmountInCents:  int(firstRow.TipAmountInCents.Int32),
		TotalPriceInCents: 0,
		UpdatedAt:         firstRow.UpdatedAt,
		Waiters:           make([]string, 0),
		Items:             make([]*dto.OrderItemDto, 0, len(rows)),
	}

	err = json.Unmarshal(firstRow.Waiters, &respDto.Waiters)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling waiters from query into array: %w", err)
	}

	if !firstRow.ItemID.Valid {
		// this means this order doesnt have any items added yet
		return respDto, nil
	}

	for _, row := range rows {
		respDto.TotalPriceInCents += int(row.PriceInCents.Int32)

		item := &dto.OrderItemDto{
			ID:           row.OrderItemID.UUID,
			RestaurantID: row.RestaurantID.UUID,
			ItemID:       row.ID,
			Name:         row.ItemName.String,
			PriceInCents: int(row.PriceInCents.Int32),
		}

		respDto.Items = append(respDto.Items, item)
	}

	return respDto, nil
}

func (r *ordersRepo) GetMenuItem(ctx context.Context, itemID uuid.UUID) (*dto.OrderItemDto, error) {
	row, err := r.q.GetMenuItem(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("fetching menu item from database: %w", err)
	}

	item := &dto.OrderItemDto{
		ID:           row.ID,
		RestaurantID: row.RestaurantID.UUID,
		ItemID:       row.ID,
		Name:         row.Name,
		PriceInCents: row.PriceInCents,
	}

	return item, nil
}

func (r *ordersRepo) DeleteOrderItem(
	ctx context.Context,
	orderItemID, orderID uuid.UUID,
) (*dto.OrderItemDto, error) {
	row, err := r.q.DeleteOrderItem(ctx, db.DeleteOrderItemParams{
		ID:      orderItemID,
		OrderID: orderID,
	})
	if err != nil {
		return nil, fmt.Errorf("deleting order item from database: %w", err)
	}

	deletedItem := &dto.OrderItemDto{
		ID:           row.ID,
		RestaurantID: uuid.Nil,
		ItemID:       row.ItemID.UUID,
		Name:         row.ItemName,
		PriceInCents: row.PriceInCents,
	}

	return deletedItem, nil
}

func (r *ordersRepo) UpdateOrder(
	ctx context.Context,
	reqDto *dto.UpdateOrderReqDto,
) (*dto.OrderDto, error) {
	var status db.OrderStatus
	if reqDto.Status != nil {
		status = *reqDto.Status
	}

	var tip int32
	if reqDto.TipAmountInCents != nil {
		tip = *reqDto.TipAmountInCents
	}

	row, err := r.q.UpdateOrder(ctx, db.UpdateOrderParams{
		ID:               reqDto.OrderID,
		Status:           db.NullOrderStatus{OrderStatus: status, Valid: reqDto.Status != nil},
		TipAmountInCents: sql.NullInt32{Int32: tip, Valid: reqDto.TipAmountInCents != nil},
	})
	if err != nil {
		return nil, fmt.Errorf("updating order in database: %w", err)
	}

	respDto := &dto.OrderDto{ //nolint:exhaustruct
		ID:               row.ID,
		Status:           row.Status,
		TipAmountInCents: int(row.TipAmountInCents.Int32),
	}

	return respDto, nil
}

func (r *ordersRepo) AssignWaiter(ctx context.Context, orderID, userID uuid.UUID) error {
	_, err := r.q.AssignWaiterToOrder(ctx, db.AssignWaiterToOrderParams{
		ID:      uuid.New(),
		UserID:  userID,
		OrderID: orderID,
	})
	if err != nil {
		return fmt.Errorf("inserting new order waiter to database: %w", err)
	}

	return nil
}

func (r *ordersRepo) RemoveWaiter(ctx context.Context, orderID, userID, assignID uuid.UUID) error {
	_, err := r.q.RemoveWaiterFromOrder(ctx, db.RemoveWaiterFromOrderParams{
		ID:      assignID,
		UserID:  userID,
		OrderID: orderID,
	})
	if err != nil {
		return fmt.Errorf("deleting order waiter from database: %w", err)
	}

	return nil
}

func (r *ordersRepo) IsUserRestaurantWaiter(
	ctx context.Context,
	userID, restaurantID uuid.UUID,
) error {
	_, err := r.q.IsUserRestaurantWaiter(ctx, db.IsUserRestaurantWaiterParams{
		UserID:       userID,
		RestaurantID: restaurantID,
	})
	if err != nil {
		return fmt.Errorf("confirming if user is restaurant manager: %w", err)
	}

	return nil
}

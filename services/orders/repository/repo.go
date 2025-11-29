// Package repository provides methods to access and manage orders data from database.
package repository

import (
	"context"
	"database/sql"
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

// Repo defines methods for accessing and managing orders data.
type Repo interface {
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
	) (uuid.UUID, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) (*dto.OrderDto, error)
	GetMenuItem(ctx context.Context, itemID uuid.UUID) (*dto.OrderItemDto, error)
}

type repo struct {
	q *db.Queries
}

// New creates a new orders reposiotry instance.
//
//revive:disable:unexported-return
func New(q *db.Queries) *repo {
	return &repo{
		q: q,
	}
}

//revive:enable:unexported-return

func (r *repo) GetCurrentOrderForTable(
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

func (r *repo) CreateOrderForTable(
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

func (r *repo) GetTableCurrency(ctx context.Context, tableID uuid.UUID) (string, error) {
	currency, err := r.q.GetTableCurrency(ctx, tableID)
	if err != nil {
		return "", fmt.Errorf("fetching table currency from database: %w", err)
	}

	return currency, nil
}

func (r *repo) AddItemToOrder(
	ctx context.Context,
	orderID uuid.UUID,
	item *dto.OrderItemDto,
) (uuid.UUID, error) {
	id, err := r.q.AddOrderItem(ctx, db.AddOrderItemParams{
		ID:           uuid.New(),
		OrderID:      orderID,
		ItemID:       uuid.NullUUID{UUID: item.ID, Valid: true},
		ItemName:     item.Name,
		PriceInCents: item.PriceInCents,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("inserting order item into database: %w", err)
	}

	return id, nil
}

func (r *repo) GetOrderItems(ctx context.Context, orderID uuid.UUID) (*dto.OrderDto, error) {
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
		Status:            string(firstRow.Status),
		Currency:          firstRow.Currency,
		TipAmountInCents:  int(firstRow.TipAmountInCents.Int32),
		TotalPriceInCents: 0,
		Items:             make([]*dto.OrderItemDto, 0, len(rows)),
	}

	if !firstRow.ItemID.Valid {
		// this means this order doesnt have any items added yet
		return respDto, nil
	}

	for _, row := range rows {
		respDto.TotalPriceInCents += int(row.PriceInCents.Int32)

		item := &dto.OrderItemDto{
			ID:           row.ID,
			RestaurantID: row.RestaurantID.UUID,
			Name:         row.ItemName.String,
			PriceInCents: int(row.PriceInCents.Int32),
		}

		respDto.Items = append(respDto.Items, item)
	}

	return respDto, nil
}

func (r *repo) GetMenuItem(ctx context.Context, itemID uuid.UUID) (*dto.OrderItemDto, error) {
	row, err := r.q.GetMenuItem(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("fetching menu item from database: %w", err)
	}

	item := &dto.OrderItemDto{
		ID:           row.ID,
		RestaurantID: row.RestaurantID.UUID,
		Name:         row.Name,
		PriceInCents: row.PriceInCents,
	}

	return item, nil
}

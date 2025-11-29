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

// ErrNoCurrentOrder is returned when there is no active order for the specified table.
var ErrNoCurrentOrder = errors.New("current order for this table doesnt exist")

// Repo defines methods for accessing and managing orders data.
type Repo interface {
	GetCurrentOrderForTable(ctx context.Context, tableID uuid.UUID) (*dto.CurrentOrderDto, error)
	CreateOrderForTable(
		ctx context.Context,
		tableID uuid.UUID,
		currency string,
	) (*dto.CurrentOrderDto, error)
	GetTableCurrency(ctx context.Context, tableID uuid.UUID) (string, error)
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

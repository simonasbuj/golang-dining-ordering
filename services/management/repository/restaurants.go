// Package repository provides methods to access and manage restaurant data.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	db "golang-dining-ordering/services/management/db/generated"
	"golang-dining-ordering/services/management/dto"

	"github.com/google/uuid"
)

// RestaurantRepository defines methods for accessing and managing restaurant data.
type RestaurantRepository interface {
	CreateRestaurant(
		ctx context.Context,
		reqDto *dto.CreateRestaurantDto,
	) (*dto.CreateRestaurantDto, error)
}

// restaurantRepository implements RestaurantRepository using sqlc-generated queries.
type restaurantRepository struct {
	db *sql.DB
	q  *db.Queries
}

// NewRestaurantRepository creates a new RestaurantRepository instance.
//
//nolint:revive // intended unexported type return
func NewRestaurantRepository(db *sql.DB, q *db.Queries) *restaurantRepository {
	return &restaurantRepository{
		db: db,
		q:  q,
	}
}

// CreateRestaurant inserts a new restaurant, adds owner to restaurant managers and returns the created DTO.
func (r *restaurantRepository) CreateRestaurant(
	ctx context.Context,
	reqDto *dto.CreateRestaurantDto,
) (*dto.CreateRestaurantDto, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	qtx := r.q.WithTx(tx)

	res, err := qtx.InsertRestaurant(ctx, db.InsertRestaurantParams{
		ID:      uuid.New().String(),
		Name:    reqDto.Name,
		Address: reqDto.Address,
	})
	if err != nil {
		_ = tx.Rollback()

		return nil, fmt.Errorf("error inserting new restaurant: %w", err)
	}

	resMngr, err := qtx.InsertRestaurantManager(ctx, db.InsertRestaurantManagerParams{
		ID:           uuid.New().String(),
		UserID:       reqDto.UserID,
		RestaurantID: res.ID,
	})
	if err != nil {
		_ = tx.Rollback()

		return nil, fmt.Errorf("error inserting new manager: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit create new restaurant transaction: %w", err)
	}

	return &dto.CreateRestaurantDto{
		ID:      res.ID,
		UserID:  resMngr.UserID,
		Name:    res.Name,
		Address: res.Address,
	}, nil
}

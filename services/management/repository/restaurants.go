// Package repository provides methods to access and manage restaurant data.
package repository

import (
	"context"
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
	q *db.Queries
}

// NewRestaurantRepository creates a new RestaurantRepository instance.
//
//nolint:revive // intended unexported type return
func NewRestaurantRepository(q *db.Queries) *restaurantRepository {
	return &restaurantRepository{
		q: q,
	}
}

// CreateRestaurant inserts a new restaurant and returns the created DTO.
func (r *restaurantRepository) CreateRestaurant(
	ctx context.Context,
	reqDto *dto.CreateRestaurantDto,
) (*dto.CreateRestaurantDto, error) {
	res, err := r.q.InsertRestaurant(ctx, db.InsertRestaurantParams{
		ID:      uuid.New().String(),
		Name:    reqDto.Name,
		Address: reqDto.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("error inserting new restaurant: %w", err)
	}

	return &dto.CreateRestaurantDto{
		ID:      res.ID,
		Name:    res.Name,
		Address: res.Address,
	}, nil
}

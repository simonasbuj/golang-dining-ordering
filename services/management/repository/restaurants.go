// Package repository provides methods to access and manage restaurant data.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	db "golang-dining-ordering/services/management/db/generated"
	"golang-dining-ordering/services/management/dto"
	"math"

	"github.com/google/uuid"
)

// RestaurantRepository defines methods for accessing and managing restaurant data.
type RestaurantRepository interface {
	CreateRestaurant(
		ctx context.Context,
		reqDto *dto.CreateRestaurantDto,
	) (*dto.CreateRestaurantDto, error)
	GetRestaurants(
		ctx context.Context,
		reqDto *dto.GetRestaurantsReqDto,
	) (*dto.GetRestaurantsRespDto, error)
	GetRestaurantByID(ctx context.Context, id uuid.UUID) (*dto.RestaurantItemDto, error)
	IsUserRestaurantManager(ctx context.Context, userID, restaurantID uuid.UUID) error
}

// restaurantRepository implements RestaurantRepository using sqlc-generated queries.
type restaurantRepository struct {
	db *sql.DB
	q  *db.Queries
}

// NewRestaurantRepository creates a new RestaurantRepository instance.
//
//revive:disable:unexported-return
func NewRestaurantRepository(db *sql.DB, q *db.Queries) *restaurantRepository {
	return &restaurantRepository{
		db: db,
		q:  q,
	}
}

//revive:enable:unexported-return

// CreateRestaurant inserts a new restaurant, adds owner to restaurant managers and returns the created DTO.
func (r *restaurantRepository) CreateRestaurant(
	ctx context.Context,
	reqDto *dto.CreateRestaurantDto,
) (*dto.CreateRestaurantDto, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("starting database transaction: %w", err)
	}

	qtx := r.q.WithTx(tx)

	res, err := qtx.InsertRestaurant(ctx, db.InsertRestaurantParams{
		ID:      uuid.New(),
		Name:    reqDto.Name,
		Address: reqDto.Address,
	})
	if err != nil {
		_ = tx.Rollback()

		return nil, fmt.Errorf("inserting new restaurant: %w", err)
	}

	resMngr, err := qtx.InsertRestaurantManager(ctx, db.InsertRestaurantManagerParams{
		ID:           uuid.New(),
		UserID:       reqDto.UserID,
		RestaurantID: res.ID,
	})
	if err != nil {
		_ = tx.Rollback()

		return nil, fmt.Errorf("inserting new manager: %w", err)
	}

	_, err = qtx.InsertRestaurantMenu(ctx, db.InsertRestaurantMenuParams{
		ID:           res.ID,
		RestaurantID: res.ID,
	})
	if err != nil {
		_ = tx.Rollback()

		return nil, fmt.Errorf("inserting restaurant menu: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("committing create new restaurant transaction: %w", err)
	}

	return &dto.CreateRestaurantDto{
		ID:      res.ID,
		UserID:  resMngr.UserID,
		Name:    res.Name,
		Address: res.Address,
	}, nil
}

func (r *restaurantRepository) GetRestaurants(
	ctx context.Context,
	reqDto *dto.GetRestaurantsReqDto,
) (*dto.GetRestaurantsRespDto, error) {
	offset := max(min((reqDto.Page-1)*reqDto.Limit, math.MaxInt32), 0)

	rows, err := r.q.GetRestaurants(ctx, db.GetRestaurantsParams{
		Limit:  reqDto.Limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("fetching restaurants with %+v: %w", reqDto, err)
	}

	respDto := &dto.GetRestaurantsRespDto{
		Page:        reqDto.Page,
		Limit:       reqDto.Limit,
		Total:       len(rows),
		Restaurants: mapGetRestaurantsRows(rows),
	}

	return respDto, nil
}

func (r *restaurantRepository) GetRestaurantByID(
	ctx context.Context,
	id uuid.UUID,
) (*dto.RestaurantItemDto, error) {
	row, err := r.q.GetRestaurantByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching restaurant with id '%s' : %w", id, err)
	}

	resDto := &dto.RestaurantItemDto{
		ID:        row.ID,
		Name:      row.Name,
		Address:   row.Address,
		CreatedAt: row.CreatedAt,
	}

	return resDto, nil
}

func (r *restaurantRepository) IsUserRestaurantManager(
	ctx context.Context,
	userID, restaurantID uuid.UUID,
) error {
	_, err := r.q.IsUserRestaurantManager(ctx, db.IsUserRestaurantManagerParams{
		UserID:       userID,
		RestaurantID: restaurantID,
	})
	if err != nil {
		return fmt.Errorf("confirming if user is restaurant manager: %w", err)
	}

	return nil
}

func mapGetRestaurantsRows(rows []db.GetRestaurantsRow) []dto.RestaurantItemDto {
	result := make([]dto.RestaurantItemDto, len(rows))
	for i, r := range rows {
		result[i] = dto.RestaurantItemDto{
			ID:        r.ID,
			Name:      r.Name,
			Address:   r.Address,
			CreatedAt: r.CreatedAt,
		}
	}

	return result
}

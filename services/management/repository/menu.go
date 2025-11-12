package repository

import (
	"context"
	"database/sql"
	"fmt"
	db "golang-dining-ordering/services/management/db/generated"
	"golang-dining-ordering/services/management/dto"

	"github.com/google/uuid"
)

// MenuRepository defines methods for accessing and managing restaurant data.
type MenuRepository interface {
	AddMenuCategory(
		ctx context.Context,
		reqDto *dto.MenuCategoryDto,
	) (*dto.MenuCategoryDto, error)
}

// menuRepository implements MenuRepository using sqlc-generated queries.
type menuRepository struct {
	db *sql.DB
	q  *db.Queries
}

// NewMenuRepository creates a new MenuRepository instance.
//
//nolint:revive // intended unexported type return
func NewMenuRepository(db *sql.DB, q *db.Queries) *menuRepository {
	return &menuRepository{
		db: db,
		q:  q,
	}
}

func (r *menuRepository) AddMenuCategory(
	ctx context.Context,
	reqDto *dto.MenuCategoryDto,
) (*dto.MenuCategoryDto, error) {
	row, err := r.q.InsertMenuCategory(ctx, db.InsertMenuCategoryParams{
		ID:          uuid.New().String(),
		MenuID:      reqDto.RestaurantID,
		Name:        reqDto.Name,
		Description: sql.NullString{String: reqDto.Description, Valid: reqDto.Description != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert new category: %w", err)
	}

	return &dto.MenuCategoryDto{
		ID:           row.ID,
		RestaurantID: "hi",
		Name:         row.Name,
		Description:  row.Description.String,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		DeletedAt:    nil,
	}, nil
}

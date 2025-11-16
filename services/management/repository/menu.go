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
	AddMenuItem(
		ctx context.Context,
		reqDto *dto.MenuItemDto,
	) (*dto.MenuItemDto, error)
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
		RestaurantID: reqDto.RestaurantID,
		Name:         row.Name,
		Description:  row.Description.String,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		DeletedAt:    nil,
	}, nil
}

func (r *menuRepository) AddMenuItem(
	ctx context.Context,
	reqDto *dto.MenuItemDto,
) (*dto.MenuItemDto, error) {
	row, err := r.q.InsertMenuItem(ctx, db.InsertMenuItemParams{
		ID:          uuid.New().String(),
		CategoryID:  reqDto.CategoryID,
		Name:        reqDto.Name,
		Description: sql.NullString{String: reqDto.Description, Valid: reqDto.Description != ""},
		Price:       reqDto.Price,
		IsAvailable: true,
		ImagePath:   sql.NullString{String: reqDto.ImagePath, Valid: reqDto.ImagePath != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert new menu item: %w", err)
	}

	return &dto.MenuItemDto{ //nolint:exhaustruct
		ID:          row.ID,
		CategoryID:  row.CategoryID,
		Name:        row.Name,
		Description: row.Description.String,
		Price:       row.Price,
		IsAvailable: row.IsAvailable,
		ImagePath:   row.ImagePath.String,
	}, nil
}

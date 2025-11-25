package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
	GetMenuItems(ctx context.Context, restaurantID uuid.UUID) (*dto.ListMenuItemsDto, error)
}

// menuRepository implements MenuRepository using sqlc-generated queries.
type menuRepository struct {
	db *sql.DB
	q  *db.Queries
}

// NewMenuRepository creates a new MenuRepository instance.
//
//revive:disable:unexported-return
func NewMenuRepository(db *sql.DB, q *db.Queries) *menuRepository {
	return &menuRepository{
		db: db,
		q:  q,
	}
}

//revive:enable:unexported-return

func (r *menuRepository) AddMenuCategory(
	ctx context.Context,
	reqDto *dto.MenuCategoryDto,
) (*dto.MenuCategoryDto, error) {
	row, err := r.q.InsertMenuCategory(ctx, db.InsertMenuCategoryParams{
		ID:          uuid.New(),
		MenuID:      reqDto.RestaurantID,
		Name:        reqDto.Name,
		Description: sql.NullString{String: reqDto.Description, Valid: reqDto.Description != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("inserting new category: %w", err)
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
		ID:           uuid.New(),
		CategoryID:   reqDto.CategoryID,
		Name:         reqDto.Name,
		Description:  sql.NullString{String: reqDto.Description, Valid: reqDto.Description != ""},
		PriceInCents: reqDto.PriceInCents,
		IsAvailable:  true,
		ImagePath:    sql.NullString{String: reqDto.ImagePath, Valid: reqDto.ImagePath != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("inserting new menu item: %w", err)
	}

	return &dto.MenuItemDto{
		ID:           row.ID,
		RestaurantID: reqDto.RestaurantID,
		CategoryID:   row.CategoryID,
		Name:         row.Name,
		Description:  row.Description.String,
		PriceInCents: row.PriceInCents,
		IsAvailable:  row.IsAvailable,
		ImagePath:    row.ImagePath.String,
		FileHeader:   nil,
	}, nil
}

func (r *menuRepository) GetMenuItems(
	ctx context.Context,
	restaurantID uuid.UUID,
) (*dto.ListMenuItemsDto, error) {
	rows, err := r.q.GetMenuCategoriesWithItems(ctx, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("fetching items from database: %w", err)
	}

	var respDto dto.ListMenuItemsDto

	err = json.Unmarshal(rows[0], &respDto)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling database results into ListMenuItemsDto: %w", err)
	}

	return &respDto, nil
}

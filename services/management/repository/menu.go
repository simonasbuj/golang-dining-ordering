package repository

import (
	"context"
	"database/sql"
	db "golang-dining-ordering/services/management/db/generated"
	"golang-dining-ordering/services/management/dto"
	"time"
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
	_ context.Context,
	_ *dto.MenuCategoryDto,
) (*dto.MenuCategoryDto, error) {
	return &dto.MenuCategoryDto{
		ID:           "hi",
		RestaurantID: "hi",
		Name:         "hi",
		Description:  "hi",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		DeletedAt:    nil,
	}, nil
}

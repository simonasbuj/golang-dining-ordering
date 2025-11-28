// Package repository provides methods to access and manage orders data from database.
package repository

import (
	"context"
	db "golang-dining-ordering/services/auth/db/generated"
	"golang-dining-ordering/services/orders/dto"

	"github.com/google/uuid"
)

type Repo interface {
	GetCurrentOrderForTable(ctx context.Context, tableID uuid.UUID) (*dto.CurrentOrderDto, error)
}

type repo struct {
	q *db.Queries
}

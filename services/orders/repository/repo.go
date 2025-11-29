// Package repository provides methods to access and manage orders data from database.
package repository

import (
	"context"
	db "golang-dining-ordering/services/auth/db/generated"
	"golang-dining-ordering/services/orders/dto"

	"github.com/google/uuid"
)

// Repo defines methods for accessing and managing orders data.
type Repo interface {
	GetCurrentOrderForTable(ctx context.Context, tableID uuid.UUID) (*dto.CurrentOrderDto, error)
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

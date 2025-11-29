// Package service provides business logic for orders creation and management.
package service

import (
	"context"
	"golang-dining-ordering/services/orders/dto"

	"github.com/google/uuid"
)

// Service defines business logic methods for orders service.
type Service interface {
	GetOrCreateCurrentOrderForTable(
		ctx context.Context,
		tableID uuid.UUID,
	) (*dto.CurrentOrderDto, error)
}

type service struct{}

// New creates a new orders service instance.
//
//revive:disable:unexported-return
func New() *service {
	return &service{}
}

//revive:enable:unexported-return

func (s *service) GetOrCreateCurrentOrderForTable(
	_ context.Context,
	tableID uuid.UUID,
) (*dto.CurrentOrderDto, error) {
	return &dto.CurrentOrderDto{
		ID: tableID,
	}, nil
}

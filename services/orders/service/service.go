// Package services provides business logic for orders creation and management.
package service

import (
	"context"
	"golang-dining-ordering/services/orders/dto"

	"github.com/google/uuid"
)

type Service interface {
	GetOrCreateCurrentOrderForTable(
		ctx context.Context,
		tableID uuid.UUID,
	) (*dto.CurrentOrderDto, error)
}

type service struct{}

func New() *service {
	return &service{}
}

func (s *service) GetOrCreateCurrentOrderForTable(
	ctx context.Context,
	tableID uuid.UUID,
) (*dto.CurrentOrderDto, error) {
	return &dto.CurrentOrderDto{
		ID: tableID,
	}, nil
}

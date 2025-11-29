// Package service provides business logic for orders creation and management.
package service

import (
	"context"
	"errors"
	"fmt"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/repository"

	"github.com/google/uuid"
)

// Service defines business logic methods for orders service.
type Service interface {
	GetOrCreateCurrentOrderForTable(
		ctx context.Context,
		tableID uuid.UUID,
	) (*dto.CurrentOrderDto, error)
}

type service struct {
	repo repository.Repo
}

// New creates a new orders service instance.
//
//revive:disable:unexported-return
func New(repo repository.Repo) *service {
	return &service{
		repo: repo,
	}
}

//revive:enable:unexported-return

func (s *service) GetOrCreateCurrentOrderForTable(
	ctx context.Context,
	tableID uuid.UUID,
) (*dto.CurrentOrderDto, error) {
	respDto, err := s.repo.GetCurrentOrderForTable(ctx, tableID)
	if err == nil {
		return respDto, nil
	}

	if !errors.Is(err, repository.ErrNoCurrentOrder) {
		return nil, fmt.Errorf("getting current order for table: %w", err)
	}

	currency, err := s.repo.GetTableCurrency(ctx, tableID)
	if err != nil {
		return nil, fmt.Errorf("getting table currency: %w", err)
	}

	respDto, err = s.repo.CreateOrderForTable(ctx, tableID, currency)
	if err != nil {
		return nil, fmt.Errorf("creating new order: %w", err)
	}

	return respDto, nil
}

// Package services provides business logic for restaurant management.
package services

import (
	"context"
	"fmt"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/repository"
)

// RestaurantService defines business logic methods for restaurants.
type RestaurantService interface {
	CreateRestaurant(
		ctx context.Context,
		reqDto *dto.CreateRestaurantDto,
	) (*dto.CreateRestaurantDto, error)
}

// restaurantService implements RestaurantService.
type restaurantService struct {
	repo repository.RestaurantRepository
}

// NewRestaurantService creates a new RestaurantService instance.
//
//nolint:revive // intended unexported type return
func NewRestaurantService(repo repository.RestaurantRepository) *restaurantService {
	return &restaurantService{
		repo: repo,
	}
}

// CreateRestaurant creates a new restaurant using the repository.
func (s *restaurantService) CreateRestaurant(
	ctx context.Context,
	reqDto *dto.CreateRestaurantDto,
) (*dto.CreateRestaurantDto, error) {
	resDto, err := s.repo.CreateRestaurant(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %w", err)
	}

	return resDto, nil
}

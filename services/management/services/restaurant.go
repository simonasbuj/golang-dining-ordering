// Package services provides business logic for restaurant management.
package services

import (
	"context"
	"errors"
	"fmt"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/repository"

	"github.com/google/uuid"
)

var errIDNotProvided = errors.New("restaurant id not provided")

// RestaurantService defines business logic methods for restaurants.
type RestaurantService interface {
	CreateRestaurant(
		ctx context.Context,
		reqDto *dto.CreateRestaurantDto,
	) (*dto.CreateRestaurantDto, error)
	GetRestaurants(
		ctx context.Context,
		reqDto *dto.GetRestaurantsReqDto,
	) (*dto.GetRestaurantsRespDto, error)
	GetRestaurantByID(ctx context.Context, id uuid.UUID) (*dto.RestaurantItemDto, error)
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
		return nil, fmt.Errorf("creating restaurant: %w", err)
	}

	return resDto, nil
}

func (s *restaurantService) GetRestaurants(
	ctx context.Context,
	reqDto *dto.GetRestaurantsReqDto,
) (*dto.GetRestaurantsRespDto, error) {
	resDto, err := s.repo.GetRestaurants(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("fetching restaurants: %w", err)
	}

	return resDto, nil
}

func (s *restaurantService) GetRestaurantByID(
	ctx context.Context,
	id uuid.UUID,
) (*dto.RestaurantItemDto, error) {
	if id == uuid.Nil {
		return nil, errIDNotProvided
	}

	resDto, err := s.repo.GetRestaurantByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching restaurant: %w", err)
	}

	return resDto, nil
}

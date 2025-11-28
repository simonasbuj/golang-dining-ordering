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
	UpdateRestaurant(
		ctx context.Context,
		reqDto *dto.UpdateRestaurantRequestDto,
	) (*dto.UpdateRestaurantResponseDto, error)
	CreateTable(
		ctx context.Context,
		reqDto *dto.RestaurantTableDto,
	) (*dto.RestaurantTableDto, error)
	GetTables(ctx context.Context, restaurantID uuid.UUID) ([]*dto.RestaurantTableDto, error)
}

// restaurantService implements RestaurantService.
type restaurantService struct {
	repo repository.RestaurantRepository
}

// NewRestaurantService creates a new RestaurantService instance.
//
//revive:disable:unexported-return
func NewRestaurantService(repo repository.RestaurantRepository) *restaurantService {
	return &restaurantService{
		repo: repo,
	}
}

//revive:enable:unexported-return

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
	if reqDto.Page == 0 {
		reqDto.Page = 1
	}

	if reqDto.Limit == 0 {
		reqDto.Limit = 10
	}

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

func (s *restaurantService) UpdateRestaurant(
	ctx context.Context,
	reqDto *dto.UpdateRestaurantRequestDto,
) (*dto.UpdateRestaurantResponseDto, error) {
	err := isUserRestaurantManager(ctx, reqDto.UserID, reqDto.ID, s.repo)
	if err != nil {
		return nil, err
	}

	respDto, err := s.repo.UpdateRestaurant(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("updating restaurant: %w", err)
	}

	return respDto, nil
}

func (s *restaurantService) CreateTable(
	ctx context.Context,
	reqDto *dto.RestaurantTableDto,
) (*dto.RestaurantTableDto, error) {
	err := isUserRestaurantManager(ctx, reqDto.UserID, reqDto.RestaurantID, s.repo)
	if err != nil {
		return nil, err
	}

	respDto, err := s.repo.CreateTable(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("creating new table: %w", err)
	}

	return respDto, nil
}

func (s *restaurantService) GetTables(
	ctx context.Context,
	restaurantID uuid.UUID,
) ([]*dto.RestaurantTableDto, error) {
	respDto, err := s.repo.GetTables(ctx, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("error fetching tables: %w", err)
	}

	return respDto, nil
}

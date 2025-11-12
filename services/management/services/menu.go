package services

import (
	"context"
	"errors"
	"fmt"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/repository"
)

const (
	userTypeManager = 1
	uerTypeWaiter   = 2
)

// ErrUserIsNotManager is returned when a user attempts an action that requires restaurant manager privileges.
var ErrUserIsNotManager = errors.New("user is not a manager")

// MenuService defines business logic methods for restaurant menus.
type MenuService interface {
	AddMenuCategory(
		ctx context.Context,
		reqDto *dto.MenuCategoryDto,
		claims *dto.TokenClaimsDto,
	) (*dto.MenuCategoryDto, error)
}

// menuService implements MenuService.
type menuService struct {
	menuRepo repository.MenuRepository
	restRepo repository.RestaurantRepository
}

// NewMenuService creates a new MenuService instance.
//
//nolint:revive // intended unexported type return
func NewMenuService(
	menuRepo repository.MenuRepository,
	restRepo repository.RestaurantRepository,
) *menuService {
	return &menuService{
		menuRepo: menuRepo,
		restRepo: restRepo,
	}
}

func (s *menuService) AddMenuCategory(
	ctx context.Context,
	reqDto *dto.MenuCategoryDto,
	claims *dto.TokenClaimsDto,
) (*dto.MenuCategoryDto, error) {
	if claims.Role != userTypeManager {
		return nil, ErrUserIsNotManager
	}

	err := s.restRepo.IsUserRestaurantManager(ctx, claims.UserID, reqDto.RestaurantID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUserIsNotManager, err)
	}

	resDto, err := s.menuRepo.AddMenuCategory(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("failed to add menu category: %w", err)
	}

	return resDto, nil
}

package services

import (
	"context"
	"errors"
	"fmt"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/repository"
	"golang-dining-ordering/services/management/storage"
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
	AddMenuItem(
		ctx context.Context,
		reqDto *dto.MenuItemDto,
		claims *dto.TokenClaimsDto,
	) (*dto.MenuItemDto, error)
}

// menuService implements MenuService.
type menuService struct {
	menuRepo repository.MenuRepository
	restRepo repository.RestaurantRepository
	storage  storage.Storage
}

// NewMenuService creates a new MenuService instance.
//
//nolint:revive // intended unexported type return
func NewMenuService(
	menuRepo repository.MenuRepository,
	restRepo repository.RestaurantRepository,
	storage  storage.Storage,
) *menuService {
	return &menuService{
		menuRepo: menuRepo,
		restRepo: restRepo,
		storage:  storage,
	}
}

func (s *menuService) AddMenuCategory(
	ctx context.Context,
	reqDto *dto.MenuCategoryDto,
	claims *dto.TokenClaimsDto,
) (*dto.MenuCategoryDto, error) {
	err := s.isUserRestaurantManager(ctx, claims, reqDto.RestaurantID)
	if err != nil {
		return nil, err
	}

	resDto, err := s.menuRepo.AddMenuCategory(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("failed to add menu category: %w", err)
	}

	return resDto, nil
}

func (s *menuService) AddMenuItem(
	ctx context.Context,
	reqDto *dto.MenuItemDto,
	claims *dto.TokenClaimsDto,
) (*dto.MenuItemDto, error) {
	err := s.isUserRestaurantManager(ctx, claims, reqDto.RestaurantID)
	if err != nil {
		return nil, err
	}

	if reqDto.FileHeader != nil {
		reqDto.ImagePath, err = s.storage.StoreMenuItemImage(reqDto.FileHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to store menu item image: %w", err)
		}
	}

	resDto, err := s.menuRepo.AddMenuItem(ctx, reqDto)
	if err != nil {
		_ = s.storage.DeleteMenuItemImage(reqDto.ImagePath)
		return nil, err
	}

	return resDto, nil
}

func (s *menuService) isUserRestaurantManager(
	ctx context.Context,
	claims *dto.TokenClaimsDto,
	restaurantID string,
) error {
	if claims.Role != userTypeManager {
		return ErrUserIsNotManager
	}

	err := s.restRepo.IsUserRestaurantManager(ctx, claims.UserID, restaurantID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUserIsNotManager, err)
	}

	return nil
}

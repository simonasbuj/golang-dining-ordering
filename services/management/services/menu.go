package services

import (
	"context"
	"fmt"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/repository"
	"golang-dining-ordering/services/management/storage"

	"github.com/google/uuid"
)

// MenuService defines business logic methods for restaurant menus.
type MenuService interface {
	AddMenuCategory(
		ctx context.Context,
		reqDto *dto.MenuCategoryDto,
		claims *authDto.TokenClaimsDto,
	) (*dto.MenuCategoryDto, error)
	AddMenuItem(
		ctx context.Context,
		reqDto *dto.MenuItemDto,
		claims *authDto.TokenClaimsDto,
	) (*dto.MenuItemDto, error)
	GetMenuItems(ctx context.Context, restaurantID uuid.UUID) (*dto.ListMenuItemsDto, error)
}

// menuService implements MenuService.
type menuService struct {
	menuRepo repository.MenuRepository
	restRepo repository.RestaurantRepository
	storage  storage.Storage
}

// NewMenuService creates a new MenuService instance.
//
//revive:disable:unexported-return
func NewMenuService(
	menuRepo repository.MenuRepository,
	restRepo repository.RestaurantRepository,
	storage storage.Storage,
) *menuService {
	return &menuService{
		menuRepo: menuRepo,
		restRepo: restRepo,
		storage:  storage,
	}
}

//revive:enable:unexported-return

func (s *menuService) AddMenuCategory(
	ctx context.Context,
	reqDto *dto.MenuCategoryDto,
	claims *authDto.TokenClaimsDto,
) (*dto.MenuCategoryDto, error) {
	err := isUserRestaurantManager(ctx, claims.UserID, reqDto.RestaurantID, s.restRepo)
	if err != nil {
		return nil, err
	}

	resDto, err := s.menuRepo.AddMenuCategory(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("adding menu category: %w", err)
	}

	return resDto, nil
}

func (s *menuService) AddMenuItem(
	ctx context.Context,
	reqDto *dto.MenuItemDto,
	claims *authDto.TokenClaimsDto,
) (*dto.MenuItemDto, error) {
	err := isUserRestaurantManager(ctx, claims.UserID, reqDto.RestaurantID, s.restRepo)
	if err != nil {
		return nil, err
	}

	if reqDto.FileHeader != nil {
		reqDto.ImagePath, err = s.storage.StoreMenuItemImage(reqDto.FileHeader)
		if err != nil {
			return nil, fmt.Errorf("storing menu item image in storage: %w", err)
		}
	}

	resDto, err := s.menuRepo.AddMenuItem(ctx, reqDto)
	if err != nil {
		_ = s.storage.DeleteMenuItemImage(reqDto.ImagePath)

		return nil, fmt.Errorf("deleting menu item's image from storage: %w", err)
	}

	return resDto, nil
}

func (s *menuService) GetMenuItems(
	ctx context.Context,
	restaurantID uuid.UUID,
) (*dto.ListMenuItemsDto, error) {
	respDto, err := s.menuRepo.GetMenuItems(ctx, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("fetching menu items: %w", err)
	}

	return respDto, nil
}

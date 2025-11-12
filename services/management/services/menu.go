package services

import (
	"context"
	"fmt"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/repository"
)

// MenuService defines business logic methods for restaurant menus.
type MenuService interface {
	AddMenuCategory(
		ctx context.Context,
		reqDto *dto.MenuCategoryDto,
	) (*dto.MenuCategoryDto, error)
}

// menuService implements MenuService.
type menuService struct {
	repo repository.MenuRepository
}

// NewMenuService creates a new MenuService instance.
//
//nolint:revive // intended unexported type return
func NewMenuService(repo repository.MenuRepository) *menuService {
	return &menuService{
		repo: repo,
	}
}

func (s *menuService) AddMenuCategory(
	ctx context.Context,
	reqDto *dto.MenuCategoryDto,
) (*dto.MenuCategoryDto, error) {
	resDto, err := s.repo.AddMenuCategory(ctx, reqDto)
	if err != nil {
		return nil, fmt.Errorf("failed to add menu category: %w", err)
	}

	return resDto, nil
}

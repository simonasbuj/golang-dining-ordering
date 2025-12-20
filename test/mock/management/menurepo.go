package management

import (
	"context"
	"golang-dining-ordering/services/management/dto"

	"github.com/google/uuid"
)

//nolint:gochecknoglobals
var (
	testCategoryID            = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testCategoryName          = "Žuvis"
	testCategoryDescription   = "Žuviška"
	testItemID                = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	testItemName              = "Menkė"
	testItemDescription       = "Pailga"
	testItemPriceInCents      = 1500
	testItemImagePath         = "uploads/uuid.jpg"
	testDifferentRestaurantID = uuid.MustParse("66666666-6666-6666-6666-666666666666")
)

type mockMenuRepo struct{}

// NewMockMenuRepo creates mock menu repo.
func NewMockMenuRepo() *mockMenuRepo { //nolint:revive
	return &mockMenuRepo{}
}

func (*mockMenuRepo) AddMenuCategory(
	_ context.Context,
	req *dto.MenuCategoryDto,
) (*dto.MenuCategoryDto, error) {
	if req.RestaurantID == testDifferentRestaurantID {
		return nil, ErrUserIsNotManager
	}

	if req.RestaurantID != testRestaurantID {
		return nil, errRepoFailed
	}

	return &dto.MenuCategoryDto{
		ID:           testCategoryID,
		RestaurantID: testRestaurantID,
		Name:         testCategoryName,
		Description:  testCategoryDescription,
		CreatedAt:    testDateTime,
		UpdatedAt:    testDateTime,
		DeletedAt:    nil,
	}, nil
}

func (*mockMenuRepo) AddMenuItem(
	_ context.Context,
	req *dto.MenuItemDto,
) (*dto.MenuItemDto, error) {
	if req.RestaurantID == testDifferentRestaurantID {
		return nil, ErrUserIsNotManager
	}

	if req.RestaurantID != testRestaurantID {
		return nil, errRepoFailed
	}

	return &dto.MenuItemDto{
		ID:           testItemID,
		RestaurantID: testRestaurantID,
		CategoryID:   testCategoryID,
		Name:         testItemName,
		Description:  testItemDescription,
		PriceInCents: testItemPriceInCents,
		IsAvailable:  true,
	}, nil
}

func (*mockMenuRepo) UpdateMenuItem(
	_ context.Context,
	req *dto.MenuItemDto,
) (*dto.MenuItemDto, error) {
	if req.RestaurantID == testDifferentRestaurantID {
		return nil, ErrUserIsNotManager
	}

	if req.RestaurantID != testRestaurantID {
		return nil, errRepoFailed
	}

	return &dto.MenuItemDto{
		ID:           testItemID,
		RestaurantID: testRestaurantID,
		CategoryID:   testCategoryID,
		Name:         testItemName,
		Description:  testItemDescription,
		PriceInCents: testItemPriceInCents,
		IsAvailable:  true,
	}, nil
}

func (*mockMenuRepo) GetMenuItems(_ context.Context, id uuid.UUID) (*dto.ListMenuItemsDto, error) {
	if id != testRestaurantID {
		return nil, errRepoFailed
	}

	return &dto.ListMenuItemsDto{
		Categories: []dto.CategoryDto{
			{
				ID:          testCategoryID,
				Name:        testCategoryName,
				Description: testCategoryDescription,
				Items: []dto.MenuItemDto{
					{
						ID:           testItemID,
						RestaurantID: testRestaurantID,
						CategoryID:   testCategoryID,
						Name:         testItemName,
						Description:  testItemDescription,
						PriceInCents: testItemPriceInCents,
						IsAvailable:  true,
					},
				},
			},
		},
	}, nil
}

func (*mockMenuRepo) GetMenuItemByID(_ context.Context, id uuid.UUID) (*dto.MenuItemDto, error) {
	if id != testItemID {
		return nil, errRepoFailed
	}

	return &dto.MenuItemDto{
		ID:           testItemID,
		RestaurantID: testRestaurantID,
		CategoryID:   testCategoryID,
		Name:         testItemName,
		Description:  testItemDescription,
		PriceInCents: testItemPriceInCents,
		IsAvailable:  true,
	}, nil
}

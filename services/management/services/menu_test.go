package services

import (
	"context"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/dto"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	mock "golang-dining-ordering/test/mock/management"
)

//nolint:gochecknoglobals
var (
	testCategoryID          = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testCategoryName        = "Žuvis"
	testCategoryDescription = "Žuviška"
	testItemID              = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	testItemName            = "Menkė"
	testItemDescription     = "Pailga"
	testItemPriceInCents    = 1500
)

type menuServiceTestSuite struct {
	suite.Suite

	svc  *menuService
	user *authDto.TokenClaimsDto
}

func (suite *menuServiceTestSuite) SetupSuite() {
	mockRestaurantsRepo := mock.NewMockRestaurantsRepo()
	mockMenuRepo := mock.NewMockMenuRepo()
	mockStorage := mock.NewMockStorage()
	suite.svc = NewMenuService(mockMenuRepo, mockRestaurantsRepo, mockStorage)

	suite.user = &authDto.TokenClaimsDto{
		UserID: testUserID,
	}
}

func TestMenuServiceTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(menuServiceTestSuite))
}

func (suite *menuServiceTestSuite) TestAddMenuCategory_Success() {
	reqDto := &dto.MenuCategoryDto{
		RestaurantID: testRestaurantID,
		Name:         testCategoryName,
		Description:  testCategoryDescription,
	}

	want := &dto.MenuCategoryDto{
		ID:           testCategoryID,
		RestaurantID: testRestaurantID,
		Name:         testCategoryName,
		Description:  testCategoryDescription,
		CreatedAt:    testDateTime,
		UpdatedAt:    testDateTime,
		DeletedAt:    nil,
	}

	got, err := suite.svc.AddMenuCategory(context.Background(), reqDto, suite.user)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *menuServiceTestSuite) TestAddMenuCategory_Error() {
	tests := []struct {
		name         string
		user         *authDto.TokenClaimsDto
		restaurantID uuid.UUID
	}{
		{"user is not a manager", &authDto.TokenClaimsDto{UserID: uuid.Nil}, testRestaurantID},
		{"repo failed", suite.user, uuid.Nil},
	}

	for _, tt := range tests {
		reqDto := &dto.MenuCategoryDto{
			RestaurantID: tt.restaurantID,
			Name:         testCategoryName,
			Description:  testCategoryDescription,
		}

		got, err := suite.svc.AddMenuCategory(context.Background(), reqDto, tt.user)
		suite.Require().Error(err)
		suite.Nil(got)
	}
}

func (suite *menuServiceTestSuite) TestAddMenuItem_Success() {
	fh := &multipart.FileHeader{
		Filename: "dummy-image.png",
		Size:     123,
	}

	reqDto := &dto.MenuItemDto{
		RestaurantID: testRestaurantID,
		CategoryID:   testCategoryID,
		Name:         testItemName,
		Description:  testItemDescription,
		PriceInCents: testItemPriceInCents,
		FileHeader:   fh,
	}

	want := &dto.MenuItemDto{
		ID:           testItemID,
		RestaurantID: testRestaurantID,
		CategoryID:   testCategoryID,
		Name:         testItemName,
		Description:  testItemDescription,
		PriceInCents: testItemPriceInCents,
		IsAvailable:  true,
	}

	got, err := suite.svc.AddMenuItem(context.Background(), reqDto, suite.user)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *menuServiceTestSuite) TestAddMenuItem_Error() {
	tests := []struct {
		name         string
		user         *authDto.TokenClaimsDto
		restaurantID uuid.UUID
		fileName     string
	}{
		{
			"user is not a manager",
			&authDto.TokenClaimsDto{UserID: uuid.Nil},
			testRestaurantID,
			"dummy-img.png",
		},
		{"repo failed", suite.user, uuid.Nil, "dummy-img.png"},
		{"repo failed", suite.user, testRestaurantID, "not-image.txt"},
	}

	for _, tt := range tests {
		fh := &multipart.FileHeader{
			Filename: tt.fileName,
			Size:     123,
		}
		reqDto := &dto.MenuItemDto{
			RestaurantID: tt.restaurantID,
			CategoryID:   testCategoryID,
			Name:         testItemName,
			Description:  testItemDescription,
			PriceInCents: testItemPriceInCents,
			FileHeader:   fh,
		}

		got, err := suite.svc.AddMenuItem(context.Background(), reqDto, tt.user)
		suite.Require().Error(err)
		suite.Nil(got)
	}
}

func (suite *menuServiceTestSuite) TestGetMenuItems_Success() {
	want := &dto.ListMenuItemsDto{
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
	}

	got, err := suite.svc.GetMenuItems(context.Background(), testRestaurantID)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *menuServiceTestSuite) TestGetMenuItems_RepoFailed() {
	got, err := suite.svc.GetMenuItems(context.Background(), uuid.Nil)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *menuServiceTestSuite) TestUpdateMenuItem_Success() {
	fh := &multipart.FileHeader{
		Filename: "dummy-image.png",
		Size:     123,
	}

	reqDto := &dto.MenuItemDto{
		ID:           testItemID,
		RestaurantID: testRestaurantID,
		CategoryID:   testCategoryID,
		Name:         testItemName,
		Description:  testItemDescription,
		PriceInCents: testItemPriceInCents,
		FileHeader:   fh,
	}

	want := &dto.MenuItemDto{
		ID:           testItemID,
		RestaurantID: testRestaurantID,
		CategoryID:   testCategoryID,
		Name:         testItemName,
		Description:  testItemDescription,
		PriceInCents: testItemPriceInCents,
		IsAvailable:  true,
	}

	got, err := suite.svc.UpdateMenuItem(context.Background(), reqDto, suite.user)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *menuServiceTestSuite) TestUpdateMenuItem_Error() {
	tests := []struct {
		name         string
		user         *authDto.TokenClaimsDto
		restaurantID uuid.UUID
		itemID       uuid.UUID
		fileName     string
	}{
		{
			"user is not a manager",
			&authDto.TokenClaimsDto{UserID: uuid.Nil},
			testRestaurantID,
			testItemID,
			"dummy-image.png",
		},
		{
			"repo failed fetching current item data",
			suite.user,
			testRestaurantID,
			uuid.Max,
			"dummy-image.png",
		},
		{"repo failed updating item", suite.user, uuid.Nil, testItemID, "dummy-image.png"},
		{"invalid file", suite.user, testRestaurantID, testItemID, "dummy-not-image.txt"},
	}

	for _, tt := range tests {
		fh := &multipart.FileHeader{
			Filename: tt.fileName,
			Size:     123,
		}
		reqDto := &dto.MenuItemDto{
			ID:           tt.itemID,
			RestaurantID: tt.restaurantID,
			CategoryID:   testCategoryID,
			Name:         testItemName,
			Description:  testItemDescription,
			PriceInCents: testItemPriceInCents,
			FileHeader:   fh,
		}

		got, err := suite.svc.UpdateMenuItem(context.Background(), reqDto, tt.user)
		suite.Require().Error(err)
		suite.Nil(got)
	}
}

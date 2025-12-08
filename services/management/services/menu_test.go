package services

import (
	"context"
	"errors"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/dto"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
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

type menuServiceTestSuite struct {
	suite.Suite

	svc  *menuService
	user *authDto.TokenClaimsDto
}

func (suite *menuServiceTestSuite) SetupSuite() {
	mockRestaurantsRepo := newMockRestaurantsRepo()
	mockMenuRepo := newMockMenuRepo()
	mockStorage := newMockStorage()
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

type mockMenuRepo struct{}

func newMockMenuRepo() *mockMenuRepo {
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

var errStorageFailed = errors.New("storage failed")

type mockStorage struct{}

func newMockStorage() *mockStorage {
	return &mockStorage{}
}

func (*mockStorage) StoreMenuItemImage(
	_ context.Context,
	fh *multipart.FileHeader,
) (string, error) {
	name := strings.ToLower(fh.Filename)

	if !strings.HasSuffix(name, ".jpg") &&
		!strings.HasSuffix(name, ".jpeg") &&
		!strings.HasSuffix(name, ".png") {
		return "", errStorageFailed
	}

	return testItemImagePath, nil
}

func (*mockStorage) DeleteMenuItemImage(_ context.Context, _ string) error {
	return nil
}

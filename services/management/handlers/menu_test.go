package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang-dining-ordering/pkg/responses"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/middleware"
	"golang-dining-ordering/services/management/services"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

type mneuHandlerTestSuite struct {
	suite.Suite

	handler     *MenuHandler
	user        *authDto.TokenClaimsDto
	invalidUser *authDto.TokenClaimsDto
}

func (suite *mneuHandlerTestSuite) SetupSuite() {
	mockMenuRepo := newMockMenuRepo()
	mockRestaurantRepo := newMockRestaurantsRepo()
	mockStorage := newMockStorage()
	svc := services.NewMenuService(mockMenuRepo, mockRestaurantRepo, mockStorage)

	suite.handler = NewMenuHandler(svc)

	suite.user = &authDto.TokenClaimsDto{
		UserID: testUserID,
	}

	suite.invalidUser = &authDto.TokenClaimsDto{
		UserID: uuid.Max,
	}
}

func TestMenuHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(mneuHandlerTestSuite))
}

func (suite *mneuHandlerTestSuite) TestHandleAddMenuCategory_Success() {
	e := echo.New()

	body := fmt.Sprintf(
		`{"name": "%s", "description": "%s"}`,
		testCategoryName,
		testCategoryDescription,
	)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	want := &responses.SuccessResponse{
		Message: "menu category created",
		Data: &dto.MenuCategoryDto{
			ID:           testCategoryID,
			RestaurantID: testRestaurantID,
			Name:         testCategoryName,
			Description:  testCategoryDescription,
			UpdatedAt:    testDateTime,
			CreatedAt:    testDateTime,
			DeletedAt:    nil,
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleAddMenuCategory(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *mneuHandlerTestSuite) TestHandleAddMenuCategory_Error() { //nolint:funlen
	e := echo.New()

	tests := []struct {
		name         string
		restaurantID string
		reqBody      string
		statusCode   int
		user         *authDto.TokenClaimsDto
	}{
		{
			"invalid restaurant id in params",
			"invalid-id",
			`{"name": "Here", "description": "Here"}`,
			http.StatusBadRequest,
			suite.user,
		},
		{
			"invalid request body",
			testRestaurantID.String(),
			`{"missing_fields": "are missing"}`,
			http.StatusBadRequest,
			suite.user,
		},
		{
			"unauthorized user",
			testDifferentRestaurantID.String(),
			`{"name": "Here", "description": "Here"}`,
			http.StatusUnauthorized,
			suite.user,
		},
		{
			"service failed",
			uuid.Max.String(),
			`{"name": "Here", "description": "Here"}`,
			http.StatusBadRequest,
			suite.user,
		},
		{
			"user missing",
			testRestaurantID.String(),
			`{"name": "Here", "description": "Here"}`,
			http.StatusBadRequest,
			&authDto.TokenClaimsDto{UserID: uuid.Nil},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body := tt.reqBody

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set(middleware.ContextKeyAuthUser, tt.user)
			c.SetParamNames(restaurantIDParamName)
			c.SetParamValues(tt.restaurantID)

			err := suite.handler.HandleAddMenuCategory(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
}

func (suite *mneuHandlerTestSuite) TestHandleAddMenuItem_Success() {
	e := echo.New()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("category_id", testCategoryID.String())
	_ = writer.WriteField("name", testItemName)
	_ = writer.WriteField("description", testItemDescription)
	_ = writer.WriteField("price_in_cents", strconv.Itoa(testItemPriceInCents))
	err := writer.Close()
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	want := &responses.SuccessResponse{
		Message: "new menu item added",
		Data: &dto.MenuItemDto{
			ID:           testItemID,
			RestaurantID: testRestaurantID,
			CategoryID:   testCategoryID,
			Name:         testItemName,
			Description:  testItemDescription,
			PriceInCents: testItemPriceInCents,
			IsAvailable:  true,
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleAddMenuItem(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *mneuHandlerTestSuite) TestHandleAddMenuItem_Error() { //nolint:funlen
	e := echo.New()

	tests := []struct {
		name         string
		restaurantID string
		categoryID   string
		statusCode   int
		user         *authDto.TokenClaimsDto
	}{
		{
			"invalid restaurant id in params",
			"invalid-id",
			testCategoryID.String(),
			http.StatusBadRequest,
			suite.user,
		},
		{
			"invalid request form",
			testRestaurantID.String(),
			"",
			http.StatusBadRequest,
			suite.user,
		},
		{
			"unauthorized user",
			testDifferentRestaurantID.String(),
			testCategoryID.String(),
			http.StatusUnauthorized,
			suite.user,
		},
		{
			"service failed",
			uuid.Max.String(),
			testCategoryID.String(),
			http.StatusBadRequest,
			suite.user,
		},
		{
			"user missing",
			testRestaurantID.String(),
			testCategoryID.String(),
			http.StatusBadRequest,
			&authDto.TokenClaimsDto{UserID: uuid.Nil},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("category_id", tt.categoryID)
			_ = writer.WriteField("name", testItemName)
			_ = writer.WriteField("description", testItemDescription)
			_ = writer.WriteField("price_in_cents", strconv.Itoa(testItemPriceInCents))
			err := writer.Close()
			suite.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set(middleware.ContextKeyAuthUser, tt.user)
			c.SetParamNames(restaurantIDParamName)
			c.SetParamValues(tt.restaurantID)

			err = suite.handler.HandleAddMenuItem(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
}

func (suite *mneuHandlerTestSuite) TestHandleUpdateMenuItem_Success() {
	e := echo.New()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("category_id", testCategoryID.String())
	_ = writer.WriteField("name", testItemName)
	_ = writer.WriteField("description", testItemDescription)
	_ = writer.WriteField("price_in_cents", strconv.Itoa(testItemPriceInCents))
	err := writer.Close()
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName, menuItemIDParamName)
	c.SetParamValues(testRestaurantID.String(), testItemID.String())

	want := &responses.SuccessResponse{
		Message: "updated menu item",
		Data: &dto.MenuItemDto{
			ID:           testItemID,
			RestaurantID: testRestaurantID,
			CategoryID:   testCategoryID,
			Name:         testItemName,
			Description:  testItemDescription,
			PriceInCents: testItemPriceInCents,
			IsAvailable:  true,
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleUpdateMenuItem(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *mneuHandlerTestSuite) TestHandleUpdateMenuItem_Error() { //nolint:funlen
	e := echo.New()

	tests := []struct {
		name         string
		restaurantID string
		itemID       string
		categoryID   string
		statusCode   int
		user         *authDto.TokenClaimsDto
	}{
		{
			"invalid restaurant id in params",
			"invalid-id",
			testItemID.String(),
			testCategoryID.String(),
			http.StatusBadRequest,
			suite.user,
		},
		{
			"invalid menu item id in params",
			testRestaurantID.String(),
			"invalid-item-id",
			testCategoryID.String(),
			http.StatusBadRequest,
			suite.user,
		},
		{
			"invalid request form",
			testRestaurantID.String(),
			testItemID.String(),
			"",
			http.StatusBadRequest,
			suite.user,
		},
		{
			"unauthorized user",
			testDifferentRestaurantID.String(),
			testItemID.String(),
			testCategoryID.String(),
			http.StatusUnauthorized,
			suite.user,
		},
		{
			"service failed",
			uuid.Max.String(),
			testItemID.String(),
			testCategoryID.String(),
			http.StatusBadRequest,
			suite.user,
		},
		{
			"user missing",
			testRestaurantID.String(),
			testItemID.String(),
			testCategoryID.String(),
			http.StatusBadRequest,
			&authDto.TokenClaimsDto{UserID: uuid.Nil},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			_ = writer.WriteField("category_id", tt.categoryID)
			_ = writer.WriteField("name", testItemName)
			_ = writer.WriteField("description", testItemDescription)
			_ = writer.WriteField("price_in_cents", strconv.Itoa(testItemPriceInCents))
			err := writer.Close()
			suite.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set(middleware.ContextKeyAuthUser, tt.user)
			c.SetParamNames(restaurantIDParamName, menuItemIDParamName)
			c.SetParamValues(tt.restaurantID, tt.itemID)

			err = suite.handler.HandleUpdateMenuItem(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
}

func (suite *mneuHandlerTestSuite) TestHandleGetMenuItems_Success() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	want := &responses.SuccessResponse{
		Message: "menu items fetched",
		Data: &dto.ListMenuItemsDto{
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
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleGetMenuItems(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *mneuHandlerTestSuite) TestHandleGetMenuItems_Error() {
	e := echo.New()

	tests := []struct {
		name         string
		restaurantID string
		statusCode   int
	}{
		{
			"invalid restaurant id in params",
			"invalid-id",
			http.StatusBadRequest,
		},
		{
			"service failed",
			uuid.Max.String(),
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames(restaurantIDParamName)
			c.SetParamValues(tt.restaurantID)

			err := suite.handler.HandleGetMenuItems(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
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
		return nil, services.ErrUserIsNotManager
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
		return nil, services.ErrUserIsNotManager
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
		return nil, services.ErrUserIsNotManager
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

func (*mockMenuRepo) GetMenuItemByID(_ context.Context, _ uuid.UUID) (*dto.MenuItemDto, error) {
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

type mockStorage struct{}

func newMockStorage() *mockStorage {
	return &mockStorage{}
}

func (*mockStorage) StoreMenuItemImage(
	_ context.Context,
	_ *multipart.FileHeader,
) (string, error) {
	return testItemImagePath, nil
}

func (*mockStorage) DeleteMenuItemImage(_ context.Context, _ string) error {
	return nil
}

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"golang-dining-ordering/pkg/responses"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/dto"
	"golang-dining-ordering/services/management/middleware"
	"golang-dining-ordering/services/management/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

//nolint:gochecknoglobals
var (
	testUserID             = uuid.MustParse("67676767-6767-6767-6767-676767676767")
	testRestaurantID       = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRestaurantCurrency = "eur"
	testRestaurantAddress  = "Miško g. 7, Raudondvaris"
	testRestaurantName     = "Viskas Viename KO"
	testTableID            = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testTableName          = "table 01"
	testTableCapacity      = 4
	testDateTime           = time.Date(2025, time.December, 5, 19, 0, 0, 0, &time.Location{})
)

type restaurantsHandlerTestSuite struct {
	suite.Suite

	handler *RestaurantsHandler
	user    *authDto.TokenClaimsDto
}

func (suite *restaurantsHandlerTestSuite) SetupSuite() {
	mockOrdersRepo := newMockRestaurantsRepo()
	svc := services.NewRestaurantService(mockOrdersRepo)

	suite.handler = NewRestaurantsHandler(svc)

	suite.user = &authDto.TokenClaimsDto{
		UserID: testUserID,
	}
}

func TestOrdersHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(restaurantsHandlerTestSuite))
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateRestaurant_Success() {
	e := echo.New()

	reqDto := &dto.CreateRestaurantDto{
		Name:     testRestaurantName,
		Address:  testRestaurantAddress,
		Currency: testRestaurantCurrency,
	}
	bodyBytes, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)

	want := &responses.SuccessResponse{
		Message: "new restaurant created",
		Data: &dto.CreateRestaurantDto{
			ID:       testRestaurantID,
			UserID:   testUserID,
			Name:     testRestaurantName,
			Address:  testRestaurantAddress,
			Currency: testRestaurantCurrency,
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleCreateRestaurant(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateRestaurant_InvalidDto() {
	e := echo.New()

	reqDto := &dto.CreateRestaurantDto{
		Name:     "",
		Address:  testRestaurantAddress,
		Currency: testRestaurantCurrency,
	}
	bodyBytes, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)

	err = suite.handler.HandleCreateRestaurant(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateRestaurant_NoUserInContext() {
	e := echo.New()

	reqDto := &dto.CreateRestaurantDto{
		Name:     testRestaurantName,
		Address:  testRestaurantAddress,
		Currency: testRestaurantCurrency,
	}
	bodyBytes, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = suite.handler.HandleCreateRestaurant(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateRestaurant_ServiceFailed() {
	e := echo.New()

	reqDto := &dto.CreateRestaurantDto{
		Name:     "fail-creation-of-this-restaurant",
		Address:  testRestaurantAddress,
		Currency: testRestaurantCurrency,
	}
	bodyBytes, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)

	err = suite.handler.HandleCreateRestaurant(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

var errRepoFailed = errors.New("repo failed")

type mockRestaurantsRepo struct{}

func newMockRestaurantsRepo() *mockRestaurantsRepo {
	return &mockRestaurantsRepo{}
}

func (*mockRestaurantsRepo) CreateRestaurant(
	_ context.Context,
	reqDto *dto.CreateRestaurantDto,
) (*dto.CreateRestaurantDto, error) {
	if reqDto.Name != testRestaurantName {
		return nil, errRepoFailed
	}

	return &dto.CreateRestaurantDto{
		ID:       testRestaurantID,
		UserID:   testUserID,
		Name:     testRestaurantName,
		Address:  testRestaurantAddress,
		Currency: testRestaurantCurrency,
	}, nil
}

func (*mockRestaurantsRepo) GetRestaurants(
	_ context.Context,
	reqDto *dto.GetRestaurantsReqDto,
) (*dto.GetRestaurantsRespDto, error) {
	return &dto.GetRestaurantsRespDto{
		Page:  reqDto.Page,
		Limit: reqDto.Limit,
		Total: 1,
		Restaurants: []dto.RestaurantItemDto{
			{
				ID:        testRestaurantID,
				Name:      testRestaurantName,
				Address:   testRestaurantAddress,
				Currency:  testRestaurantCurrency,
				CreatedAt: testDateTime,
			},
		},
	}, nil
}

func (*mockRestaurantsRepo) GetRestaurantByID(
	_ context.Context,
	_ uuid.UUID,
) (*dto.RestaurantItemDto, error) {
	return &dto.RestaurantItemDto{
		ID:        testRestaurantID,
		Name:      testRestaurantName,
		Address:   testRestaurantAddress,
		Currency:  testRestaurantCurrency,
		CreatedAt: testDateTime,
	}, nil
}

func (*mockRestaurantsRepo) IsUserRestaurantManager(
	_ context.Context,
	_, _ uuid.UUID,
) error {
	return nil
}

func (*mockRestaurantsRepo) UpdateRestaurant(
	_ context.Context,
	_ *dto.UpdateRestaurantRequestDto,
) (*dto.UpdateRestaurantResponseDto, error) {
	return &dto.UpdateRestaurantResponseDto{
		ID:        testRestaurantID,
		Name:      testRestaurantName,
		Address:   testRestaurantAddress,
		Currency:  testRestaurantCurrency,
		CreatedAt: testDateTime,
		UpdatedAt: testDateTime,
		DeletedAt: testDateTime,
	}, nil
}

func (*mockRestaurantsRepo) CreateTable(
	_ context.Context,
	_ *dto.RestaurantTableDto,
) (*dto.RestaurantTableDto, error) {
	return &dto.RestaurantTableDto{
		ID:           testTableID,
		RestaurantID: testRestaurantID,
		UserID:       testUserID,
		Name:         testTableName,
		Capacity:     testTableCapacity,
	}, nil
}

func (*mockRestaurantsRepo) GetTables(
	_ context.Context,
	_ uuid.UUID,
) ([]*dto.RestaurantTableDto, error) {
	return []*dto.RestaurantTableDto{
		{
			ID:           testTableID,
			RestaurantID: testRestaurantID,
			UserID:       testUserID,
			Name:         testTableName,
			Capacity:     testTableCapacity,
		},
	}, nil
}

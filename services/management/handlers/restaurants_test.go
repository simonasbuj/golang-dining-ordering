package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	mock "golang-dining-ordering/test/mock/management"
)

//nolint:gochecknoglobals
var (
	testUserID               = uuid.MustParse("67676767-6767-6767-6767-676767676767")
	testRestaurantID         = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRestaurantCurrency   = "eur"
	testRestaurantAddress    = "Miško g. 7, Raudondvaris"
	testRestaurantName       = "Viskas Viename KO"
	testTableID              = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testTableName            = "table 01"
	testTableCapacity        = 4
	testDateTime             = time.Date(2025, time.December, 5, 19, 0, 0, 0, &time.Location{})
	testCreateRestaurantBody = `{"name": "name"}`
	testCreateTableBody      = `{"capacity": 4, "name": "table 01"}`
)

type restaurantsHandlerTestSuite struct {
	suite.Suite

	handler *RestaurantsHandler
	user    *authDto.TokenClaimsDto
}

func (suite *restaurantsHandlerTestSuite) SetupSuite() {
	mockOrdersRepo := mock.NewMockRestaurantsRepo()
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

func (suite *restaurantsHandlerTestSuite) TestHandleCreateRestaurant_Error() {
	e := echo.New()

	tests := []struct {
		name           string
		restaurantName string
		userContextKey string
		statusCode     int
	}{
		{
			name:           "invalid dto",
			restaurantName: "",
			userContextKey: middleware.ContextKeyAuthUser,
			statusCode:     http.StatusBadRequest,
		},
		{
			name:           "no user in context",
			restaurantName: testRestaurantName,
			userContextKey: "fakeKey",
			statusCode:     http.StatusBadRequest,
		},
		{
			name:           "service failed",
			restaurantName: "fail-this-restaurant",
			userContextKey: middleware.ContextKeyAuthUser,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reqDto := &dto.CreateRestaurantDto{
				Name:     tt.restaurantName,
				Address:  testRestaurantAddress,
				Currency: testRestaurantCurrency,
			}
			bodyBytes, err := json.Marshal(reqDto)
			suite.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set(tt.userContextKey, suite.user)

			err = suite.handler.HandleCreateRestaurant(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetRestaurants_Success() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	want := &responses.SuccessResponse{
		Message: "restaurants fetched",
		Data: &dto.GetRestaurantsRespDto{
			Page:  1,
			Limit: 10,
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
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleGetRestaurants(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetRestaurants_Error() {
	e := echo.New()

	tests := []struct {
		name  string
		page  string
		limit string
	}{
		{"invalid limit", "1", "1000"},
		{"service failed", "69", "10"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/?page=%s&limit=%s", tt.page, tt.limit),
				nil,
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := suite.handler.HandleGetRestaurants(c)
			suite.Require().Error(err)
			suite.Equal(http.StatusBadRequest, rec.Code)
		})
	}
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetRestaurantByID_Success() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	want := &responses.SuccessResponse{
		Message: "restaurant fetched",
		Data: &dto.RestaurantItemDto{
			ID:        testRestaurantID,
			Name:      testRestaurantName,
			Address:   testRestaurantAddress,
			Currency:  testRestaurantCurrency,
			CreatedAt: testDateTime,
		},
	}

	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleGetRestaurantByID(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetRestaurantByID_Error() {
	e := echo.New()

	tests := []struct {
		name         string
		restaurantID string
	}{
		{"invalid url params", "invalid-id"},
		{"service failed", uuid.Nil.String()},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames(restaurantIDParamName)
			c.SetParamValues(tt.restaurantID)

			err := suite.handler.HandleGetRestaurantByID(c)
			suite.Require().Error(err)
			suite.Equal(http.StatusBadRequest, rec.Code)
		})
	}
}

func (suite *restaurantsHandlerTestSuite) TestHandleUpdateRestaurant_Success() {
	e := echo.New()

	body := testCreateRestaurantBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	want := &responses.SuccessResponse{
		Message: "restaurant updated",
		Data: &dto.UpdateRestaurantResponseDto{
			ID:        testRestaurantID,
			Name:      testRestaurantName,
			Address:   testRestaurantAddress,
			Currency:  testRestaurantCurrency,
			CreatedAt: testDateTime,
			UpdatedAt: testDateTime,
			DeletedAt: testDateTime,
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleUpdateRestaurant(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *restaurantsHandlerTestSuite) TestHandleUpdateRestaurant_Error() {
	e := echo.New()

	tests := []struct {
		name           string
		body           string
		restaurantID   string
		userContextKey string
	}{
		{
			name:           "invalid id in params",
			body:           testCreateRestaurantBody,
			restaurantID:   "invalid-id",
			userContextKey: middleware.ContextKeyAuthUser,
		},
		{
			name:           "no user in context",
			body:           testCreateRestaurantBody,
			restaurantID:   testRestaurantID.String(),
			userContextKey: "wrong-context-key",
		},
		{
			name:           "invalid dto",
			body:           `{"user_id": "", "name": "name"}`,
			restaurantID:   testRestaurantID.String(),
			userContextKey: middleware.ContextKeyAuthUser,
		},
		{
			name:           "service failed",
			body:           testCreateRestaurantBody,
			restaurantID:   uuid.Max.String(),
			userContextKey: middleware.ContextKeyAuthUser,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set(tt.userContextKey, suite.user)
			c.SetParamNames(restaurantIDParamName)
			c.SetParamValues(tt.restaurantID)

			err := suite.handler.HandleUpdateRestaurant(c)
			suite.Require().Error(err)
			suite.Equal(http.StatusBadRequest, rec.Code)
		})
	}
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateTable_Success() {
	e := echo.New()

	body := testCreateTableBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	want := &responses.SuccessResponse{
		Message: "table added to restaurant",
		Data: &dto.RestaurantTableDto{
			ID:           testTableID,
			RestaurantID: testRestaurantID,
			UserID:       testUserID,
			Capacity:     testTableCapacity,
			Name:         testTableName,
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleCreateTable(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateTable_Error() { //nolint:funlen
	e := echo.New()

	tests := []struct {
		name           string
		body           string
		restaurantID   string
		userContextKey string
		user           *authDto.TokenClaimsDto
		statusCode     int
	}{
		{
			name:           "invalid id in params",
			body:           testCreateTableBody,
			restaurantID:   "invalid-id",
			userContextKey: middleware.ContextKeyAuthUser,
			user:           &authDto.TokenClaimsDto{UserID: testUserID},
			statusCode:     http.StatusBadRequest,
		},
		{
			name:           "no user in context",
			body:           testCreateTableBody,
			restaurantID:   testRestaurantID.String(),
			userContextKey: "wrong-context-key",
			user:           &authDto.TokenClaimsDto{UserID: testUserID},
			statusCode:     http.StatusBadRequest,
		},
		{
			name:           "invalid dto",
			body:           `{"missing_field": "required fields are not in this json"}`,
			restaurantID:   testRestaurantID.String(),
			userContextKey: middleware.ContextKeyAuthUser,
			user:           &authDto.TokenClaimsDto{UserID: testUserID},
			statusCode:     http.StatusBadRequest,
		},
		{
			name:           "user is not a manager",
			body:           testCreateTableBody,
			restaurantID:   uuid.Max.String(),
			userContextKey: middleware.ContextKeyAuthUser,
			user:           &authDto.TokenClaimsDto{UserID: uuid.Max},
			statusCode:     http.StatusUnauthorized,
		},
		{
			name:           "service failed",
			body:           testCreateTableBody,
			restaurantID:   uuid.Max.String(),
			userContextKey: middleware.ContextKeyAuthUser,
			user:           &authDto.TokenClaimsDto{UserID: testUserID},
			statusCode:     http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.Set(tt.userContextKey, tt.user)
			c.SetParamNames(restaurantIDParamName)
			c.SetParamValues(tt.restaurantID)

			err := suite.handler.HandleCreateTable(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetTables_Success() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	want := &responses.SuccessResponse{
		Message: "tables fetched",
		Data: []*dto.RestaurantTableDto{
			{
				ID:           testTableID,
				RestaurantID: testRestaurantID,
				UserID:       testUserID,
				Name:         testTableName,
				Capacity:     testTableCapacity,
			},
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleGetTables(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetTables_Error() {
	e := echo.New()

	tests := []struct {
		name         string
		restaurantID string
	}{
		{"invalid restaurant id in params", "invalid-id"},
		{"service failed", uuid.Max.String()},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames(restaurantIDParamName)
			c.SetParamValues(tt.restaurantID)

			err := suite.handler.HandleGetTables(c)
			suite.Require().Error(err)
			suite.Equal(http.StatusBadRequest, rec.Code)
		})
	}
}

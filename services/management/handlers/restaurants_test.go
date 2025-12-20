package handlers

import (
	"bytes"
	"encoding/json"
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

func (suite *restaurantsHandlerTestSuite) TestHandleGetRestaurants_InvalidLimitInDto() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/?page=1&limit=10000", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := suite.handler.HandleGetRestaurants(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetRestaurants_ServiceFailed() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/?page=69&limit=10", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := suite.handler.HandleGetRestaurants(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
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

func (suite *restaurantsHandlerTestSuite) TestHandleGetRestaurantByID_InvalidRestaurantIDInParams() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("invalid-id")
	c.SetParamValues(testRestaurantID.String())

	err := suite.handler.HandleGetRestaurantByID(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetRestaurantByID_ServiceFailed() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(uuid.Nil.String())

	err := suite.handler.HandleGetRestaurantByID(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
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

func (suite *restaurantsHandlerTestSuite) TestHandleUpdateRestaurant_InvalidRestaurantIDInParams() {
	e := echo.New()

	body := testCreateRestaurantBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues("invalid-id")

	err := suite.handler.HandleUpdateRestaurant(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleUpdateRestaurant_NoUserInContext() {
	e := echo.New()

	body := testCreateRestaurantBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	err := suite.handler.HandleUpdateRestaurant(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleUpdateRestaurant_InvalidDto() {
	e := echo.New()

	body := `{"user_id": "", "name": "name"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	err := suite.handler.HandleUpdateRestaurant(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleUpdateRestaurant_ServiceFailed() {
	e := echo.New()

	body := testCreateRestaurantBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(uuid.Max.String())

	err := suite.handler.HandleUpdateRestaurant(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
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

func (suite *restaurantsHandlerTestSuite) TestHandleCreateTable_InvalidRestaurantIDInParams() {
	e := echo.New()

	body := testCreateTableBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues("invalid-id")

	err := suite.handler.HandleCreateTable(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateTable_InvalidDto() {
	e := echo.New()

	body := `{"missing_field": "required fields are not in this json"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	err := suite.handler.HandleCreateTable(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateTable_MissingUser() {
	e := echo.New()

	body := testCreateTableBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	err := suite.handler.HandleCreateTable(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateTable_UserIsNotAManager() {
	e := echo.New()

	body := testCreateTableBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	wrongUser := *suite.user
	wrongUser.UserID = uuid.Max

	c.Set(middleware.ContextKeyAuthUser, &wrongUser)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(testRestaurantID.String())

	err := suite.handler.HandleCreateTable(c)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, services.ErrUserIsNotManager)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleCreateTable_ServiceFailed() {
	e := echo.New()

	body := testCreateTableBody
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set(middleware.ContextKeyAuthUser, suite.user)
	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(uuid.Max.String())

	err := suite.handler.HandleCreateTable(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
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

func (suite *restaurantsHandlerTestSuite) TestHandleGetTables_InvalidRestaurantIDInParams() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues("invalid-id")

	err := suite.handler.HandleGetTables(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *restaurantsHandlerTestSuite) TestHandleGetTables_ServiceFailed() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(restaurantIDParamName)
	c.SetParamValues(uuid.Max.String())

	err := suite.handler.HandleGetTables(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

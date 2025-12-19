package handlers

import (
	"bytes"
	"encoding/json"
	"golang-dining-ordering/pkg/responses"
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/middleware"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/services"
	"net/http"
	"net/http/httptest"
	"testing"

	mock "golang-dining-ordering/test/mock/orders"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ordersHandlerTestSuite struct {
	suite.Suite

	handler *OrdersHandler
	order   dto.OrderDto
}

func (suite *ordersHandlerTestSuite) SetupSuite() {
	mockOrdersRepo := mock.NewMockOrdersRepo()
	svc := services.NewOrdersService(mockOrdersRepo)

	suite.handler = NewOrdersHandler(svc)

	suite.order = dto.OrderDto{
		ID:                testOrderID,
		RestaurantID:      testRestaurantID,
		RestaurantName:    testRestaurantName,
		Status:            db.OrderStatusOpen,
		Currency:          testCurrency,
		TipAmountInCents:  testAmount,
		TotalPriceInCents: testAmount,
		UpdatedAt:         testDateTime,
		Items: []*dto.OrderItemDto{
			{
				ID:           testOrderItemID,
				RestaurantID: testRestaurantID,
				ItemID:       testItemID,
				Name:         testItemName,
				PriceInCents: testAmount,
			},
		},
	}
}

func TestOrdersHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ordersHandlerTestSuite))
}

func (suite *ordersHandlerTestSuite) TestHandleGetCurrentTableOrder_Success() {
	e := echo.New()

	target := "/?tableId=" + testTableID.String()
	req := httptest.NewRequest(http.MethodGet, target, nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	want := responses.SuccessResponse{
		Message: "fetched current order",
		Data: dto.CurrentOrderDto{
			ID: testOrderID,
		},
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleGetCurrentTableOrder(c)
	suite.Require().NoError(err)
	suite.JSONEq(string(wantJSON), rec.Body.String())
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleGetCurrentTableOrder_FailedToParseTableIDFromUrl() {
	e := echo.New()

	target := "/?tableId=this-is-not-uuid"
	req := httptest.NewRequest(http.MethodGet, target, nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := suite.handler.HandleGetCurrentTableOrder(c)
	suite.Require().Error(err)
	suite.Contains(rec.Body.String(), "failed to parse tableId from url")
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleGetCurrentTableOrder_ServiceError() {
	e := echo.New()

	target := "/?tableId=" + uuid.Max.String()
	req := httptest.NewRequest(http.MethodGet, target, nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := suite.handler.HandleGetCurrentTableOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleGetOrder_Success() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	want := responses.SuccessResponse{
		Message: "fetched order details",
		Data:    &suite.order,
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleGetOrder(c)
	suite.Require().NoError(err)
	suite.JSONEq(string(wantJSON), rec.Body.String())
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleGetOrder_InvalidOrderIDParam() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues("invalid-id")

	err := suite.handler.HandleGetOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleGetOrder_ServiceFailed() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(uuid.Max.String())

	err := suite.handler.HandleGetOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleAddItemToOrder_Success() {
	e := echo.New()

	reqDto := &dto.OrderItemRequestDto{
		ItemID: testItemID,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	updatedOrder := suite.order
	updatedOrder.Items = append(updatedOrder.Items, &dto.OrderItemDto{
		ID:           testOrderItemID,
		RestaurantID: testRestaurantID,
		ItemID:       testItemID,
		Name:         testItemName,
		PriceInCents: testAmount,
	})
	updatedOrder.TotalPriceInCents += 10
	want := &responses.SuccessResponse{
		Message: "item added to order",
		Data:    &updatedOrder,
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleAddItemToOrder(c)
	suite.Require().NoError(err)
	suite.JSONEq(string(wantJSON), rec.Body.String())
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleAddItemToOrder_InvalidOrderIDParam() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues("invalid-id")

	err := suite.handler.HandleAddItemToOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleAddItemToOrder_InvalidReqDto() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	err := suite.handler.HandleAddItemToOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleAddItemToOrder_CantAddToCompletedOrder() {
	e := echo.New()

	reqDto := &dto.OrderItemRequestDto{
		ItemID: testItemID,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testCompletedOrderID.String())

	err = suite.handler.HandleAddItemToOrder(c)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, services.ErrOrderIsNotOpen)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleAddItemToOrder_ServiceFailed() {
	e := echo.New()

	reqDto := &dto.OrderItemRequestDto{
		ItemID: testItemID,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(uuid.Max.String())

	err = suite.handler.HandleAddItemToOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleDeleteItemFromOrder_Success() {
	e := echo.New()

	reqDto := &dto.OrderItemRequestDto{
		ItemID: testOrderItemID,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	updatedOrder := suite.order
	updatedOrder.Items = []*dto.OrderItemDto{}
	updatedOrder.TotalPriceInCents = 0
	want := &responses.SuccessResponse{
		Message: "deleted item from order",
		Data:    &updatedOrder,
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleDeleteItemFromOrder(c)
	suite.Require().NoError(err)
	suite.JSONEq(string(wantJSON), rec.Body.String())
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleDeleteItemFromOrder_InvalidOrderIDParam() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues("invalid-id")

	err := suite.handler.HandleDeleteItemFromOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleDeleteItemFromOrder_InvalidReqDto() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	err := suite.handler.HandleDeleteItemFromOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleDeleteItemFromOrder_CantDeleteFromCompletedOrder() {
	e := echo.New()

	reqDto := &dto.OrderItemRequestDto{
		ItemID: testItemID,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testCompletedOrderID.String())

	err = suite.handler.HandleDeleteItemFromOrder(c)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, services.ErrOrderIsNotOpen)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleDeleteItemFromOrder_ServiceFailed() {
	e := echo.New()

	reqDto := &dto.OrderItemRequestDto{
		ItemID: testItemID,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(uuid.Max.String())

	err = suite.handler.HandleDeleteItemFromOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleUpdateOrder_Success() {
	e := echo.New()

	amount := int32(testAmount) //nolint: gosec
	status := db.OrderStatusLocked
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: &amount,
		Status:           &status,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
		UserID: testUserID,
	})

	udpatedOrder := &suite.order
	udpatedOrder.Status = db.OrderStatusLocked
	want := &responses.SuccessResponse{
		Message: "updated order",
		Data:    udpatedOrder,
	}
	wantJSON, err := json.Marshal(want)
	suite.Require().NoError(err)

	err = suite.handler.HandleUpdateOrder(c)
	suite.Require().NoError(err)
	suite.JSONEq(string(wantJSON), rec.Body.String())
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleUpdateOrder_InvalidOrderIDParam() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues("invalid-id")

	err := suite.handler.HandleUpdateOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleUpdateOrder_InvalidReqDto() {
	e := echo.New()

	amount := int32(50000000)
	status := db.OrderStatusLocked
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: &amount,
		Status:           &status,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
		UserID: testUserID,
	})

	err = suite.handler.HandleUpdateOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleUpdateOrder_OrderFinalized() {
	e := echo.New()

	amount := int32(testAmount) //nolint: gosec
	status := db.OrderStatusLocked
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testCompletedOrderID,
		TipAmountInCents: &amount,
		Status:           &status,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testCompletedOrderID.String())

	c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
		UserID: testUserID,
	})

	err = suite.handler.HandleUpdateOrder(c)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, services.ErrOrderFinalized)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleUpdateOrder_ServiceFailed() {
	e := echo.New()

	amount := int32(testAmount) //nolint: gosec
	status := db.OrderStatusLocked
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          uuid.Max,
		TipAmountInCents: &amount,
		Status:           &status,
	}

	body, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(uuid.Max.String())

	c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
		UserID: testUserID,
	})

	err = suite.handler.HandleUpdateOrder(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleAddWaiter_Success() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
		UserID: testUserID,
	})

	err := suite.handler.HandleAddWaiter(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleAddWaiter_Error() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	tests := []struct {
		name    string
		orderID string
		userID  uuid.UUID
	}{
		{"invalid order id in param", "invalid-id", testUserID},
		{"no user in context", testOrderID.String(), uuid.Nil},
		{"service failed", uuid.Nil.String(), testUserID},
	}

	for _, tt := range tests {
		c.SetParamNames(orderIDParamName)
		c.SetParamValues(tt.orderID)

		c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
			UserID: tt.userID,
		})

		err := suite.handler.HandleAddWaiter(c)
		suite.Require().Error(err)
		suite.Equal(http.StatusBadRequest, rec.Code)
	}
}

func (suite *ordersHandlerTestSuite) TestHandleRemovedWaiter_Success() {
	e := echo.New()

	body := `{"assign_id": "67676767-6767-4676-8767-676767676767"}`

	req := httptest.NewRequest(http.MethodDelete, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
		UserID: testUserID,
	})

	err := suite.handler.HandleRemoveWaiter(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *ordersHandlerTestSuite) TestHandleRemoveWaiter_Error() {
	e := echo.New()

	body := `{"assign_id": "67676767-6767-4676-8767-676767676767"}`

	tests := []struct {
		name    string
		orderID string
		userID  uuid.UUID
		body    string
	}{
		{"invalid order id in param", "invalid-id", testUserID, body},
		{"no user in context", testOrderID.String(), uuid.Nil, body},
		{"invalid request body", testOrderID.String(), testUserID, "zzz"},
		{"service failed", uuid.Nil.String(), testUserID, body},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodDelete, "/", bytes.NewReader([]byte(tt.body)))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.SetParamNames(orderIDParamName)
		c.SetParamValues(tt.orderID)

		c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
			UserID: tt.userID,
		})

		err := suite.handler.HandleRemoveWaiter(c)
		suite.Require().Error(err)
		suite.Equal(http.StatusBadRequest, rec.Code)
	}
}

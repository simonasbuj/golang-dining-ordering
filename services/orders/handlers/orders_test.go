package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (suite *ordersHandlerTestSuite) TestHandleGetCurrentTableOrder_Error() {
	e := echo.New()

	tests := []struct {
		desc       string
		targetURL  string
		statusCode int
	}{
		{"invalid table id in url params", "/?tableId=this-is-not-uuid", http.StatusBadRequest},
		{"service error", "/?tableId=" + uuid.Max.String(), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		suite.T().Run(tt.desc, func(_ *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.targetURL, nil)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := suite.handler.HandleGetCurrentTableOrder(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
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

func (suite *ordersHandlerTestSuite) TestHandleGetOrder_Error() {
	e := echo.New()

	tests := []struct {
		desc       string
		orderID    string
		statusCode int
	}{
		{"invalid id in url params", "invalid-id", http.StatusBadRequest},
		{"service error", uuid.Max.String(), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		suite.T().Run(tt.desc, func(_ *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames(orderIDParamName)
			c.SetParamValues(tt.orderID)

			err := suite.handler.HandleGetOrder(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
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

func (suite *ordersHandlerTestSuite) TestHandleAddItemToOrder_Error() {
	e := echo.New()

	tests := []struct {
		desc       string
		orderID    string
		itemID     string
		statusCode int
	}{
		{"invalid id in params", "invalid-id", testItemID.String(), http.StatusBadRequest},
		{"invalid dto", testOrderID.String(), "", http.StatusBadRequest},
		{
			"cant add to completed order",
			testCompletedOrderID.String(),
			testItemID.String(),
			http.StatusBadRequest,
		},
		{"service error", uuid.Max.String(), testItemID.String(), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		suite.T().Run(tt.desc, func(_ *testing.T) {
			body := fmt.Sprintf(`{"item_id": "%s"}`, tt.itemID)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames(orderIDParamName)
			c.SetParamValues(tt.orderID)

			err := suite.handler.HandleAddItemToOrder(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
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

func (suite *ordersHandlerTestSuite) TestDeleteItemFromOrder_Error() {
	e := echo.New()

	tests := []struct {
		desc       string
		orderID    string
		itemID     string
		statusCode int
	}{
		{"invalid id in params", "invalid-id", testItemID.String(), http.StatusBadRequest},
		{"invalid dto", testOrderID.String(), "", http.StatusBadRequest},
		{
			"cant add to completed order",
			testCompletedOrderID.String(),
			testItemID.String(),
			http.StatusBadRequest,
		},
		{"service error", uuid.Max.String(), testItemID.String(), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		suite.T().Run(tt.desc, func(_ *testing.T) {
			body := fmt.Sprintf(`{"item_id": "%s"}`, tt.itemID)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames(orderIDParamName)
			c.SetParamValues(tt.orderID)

			err := suite.handler.HandleDeleteItemFromOrder(c)
			suite.Require().Error(err)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
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

func (suite *ordersHandlerTestSuite) TestHandleUpdateOrder_Error() {
	e := echo.New()

	tests := []struct {
		desc       string
		orderID    string
		tipAmount  int
		statusCode int
	}{
		{"invalid id in params", "invalid-id", testAmount, http.StatusBadRequest},
		{"invalid dto", testOrderID.String(), 50000000, http.StatusBadRequest},
		{
			"cant update finalized order",
			testCompletedOrderID.String(),
			testAmount,
			http.StatusBadRequest,
		},
		{"service error", uuid.New().String(), testAmount, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		suite.T().Run(tt.desc, func(_ *testing.T) {
			body := fmt.Sprintf(`{"tip_amount_in_cents": %d, "status": "locked"}`, tt.tipAmount)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames(orderIDParamName)
			c.SetParamValues(tt.orderID)

			c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
				UserID: testUserID,
			})

			err := suite.handler.HandleUpdateOrder(c)
			suite.Require().Error(err)
			// suite.Require().ErrorIs(err, services.ErrOrderFinalized)
			suite.Equal(tt.statusCode, rec.Code)
		})
	}
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
		suite.T().Run(tt.name, func(_ *testing.T) {
			c.SetParamNames(orderIDParamName)
			c.SetParamValues(tt.orderID)

			c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
				UserID: tt.userID,
			})

			err := suite.handler.HandleAddWaiter(c)
			suite.Require().Error(err)
			suite.Equal(http.StatusBadRequest, rec.Code)
		})
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
		suite.T().Run(tt.name, func(_ *testing.T) {
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
		})
	}
}

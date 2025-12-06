package handlers

import (
	"encoding/json"
	"golang-dining-ordering/pkg/responses"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ordersHandlerTestSuite struct {
	suite.Suite

	handler *OrdersHandler
}

func (suite *ordersHandlerTestSuite) SetupSuite() {
	mockOrdersRepo := NewMockOrdersRepo()
	svc := services.NewOrdersService(mockOrdersRepo)

	suite.handler = NewOrdersHandler(svc)
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

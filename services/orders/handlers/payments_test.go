package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"golang-dining-ordering/pkg/responses"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mock "golang-dining-ordering/test/mock/orders"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

//nolint:gochecknoglobals
var (
	testUserID            = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	testRestaurantID      = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRestaurantName    = "Test Restaurant"
	testCurrency          = "eur"
	testAmount            = 10
	testPaymentID         = uuid.MustParse("67676767-6767-4676-8767-676767676767")
	testOrderID           = uuid.MustParse("99999999-9999-4999-9999-999999999999")
	testCompletedOrderID  = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	testTableID           = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testDateTime          = time.Date(2025, time.December, 5, 19, 0, 0, 0, &time.Location{})
	testItemName          = "Test Menu Item"
	testOrderItemID       = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	testItemID            = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	testCheckoutURL       = "http://fake-checkout-session.com/1"
	testPaymentProvider   = db.OrdersPaymentProviderMock
	testProviderPaymentID = "pi_123456"
)

type paymentsHandlerTestSuite struct {
	suite.Suite

	handler *PaymentsHandler
}

func (suite *paymentsHandlerTestSuite) SetupSuite() {
	mockOrdersRepo := mock.NewMockOrdersRepo()
	mockPaymentsRepo := mock.NewMockPaymentsRepo()
	mockPaymentsProvider := mock.NewMockPaymentsProvider()
	svc := services.NewPaymentsService(mockOrdersRepo, mockPaymentsRepo, mockPaymentsProvider)

	suite.handler = NewPaymentsHandler(svc)
}

func TestPaymentsHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(paymentsHandlerTestSuite))
}

func (suite *paymentsHandlerTestSuite) TestHandleCreateCheckout_Success() {
	e := echo.New()

	reqDto := dto.CheckoutSessionRequestDto{
		OrderDto:   nil,
		SuccessURL: "https://fake-website.io?success=true",
		CancelURL:  "https://fake-website.io?cancel=true",
	}
	bodyBytes, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	expectedResponseDto := responses.SuccessResponse{
		Message: "checkout session created",
		Data: &dto.CheckoutSessionResponseDto{
			URL:      testCheckoutURL,
			Provider: testPaymentProvider,
		},
	}
	expectedJSON, err := json.Marshal(expectedResponseDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	err = suite.handler.HandleCreateCheckout(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(expectedJSON), rec.Body.String())
}

func (suite *paymentsHandlerTestSuite) TestHandleCreateCheckout_InvalidDto() {
	e := echo.New()

	reqDto := dto.CheckoutSessionRequestDto{
		OrderDto:   nil,
		SuccessURL: "",
		CancelURL:  "",
	}
	bodyBytes, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(testOrderID.String())

	err = suite.handler.HandleCreateCheckout(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *paymentsHandlerTestSuite) TestHandleCreateCheckout_InvalidParam() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := suite.handler.HandleCreateCheckout(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

func (suite *paymentsHandlerTestSuite) TestHandleCreateCheckout_ServiceError() {
	e := echo.New()

	reqDto := dto.CheckoutSessionRequestDto{
		OrderDto:   nil,
		SuccessURL: "https://fake-website.io?success=true",
		CancelURL:  "https://fake-website.io?cancel=true",
	}
	bodyBytes, err := json.Marshal(reqDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames(orderIDParamName)
	c.SetParamValues(uuid.Max.String())

	err = suite.handler.HandleCreateCheckout(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusInternalServerError, rec.Code)
}

// func (suite *paymentsHandlerTestSuite) TestHandleCreateCheckout_Error() {
// 	e := echo.New()

// }

func (suite *paymentsHandlerTestSuite) TestHandleWebhookSuccess_Success() {
	e := echo.New()

	payload := []byte(`{"payment_secret": "secret"}`)

	wantDto := responses.SuccessResponse{
		Message: "payment verified and saved",
		Data: &dto.PaymentDto{
			ID:                testPaymentID,
			OrderID:           testOrderID,
			AmountInCents:     testAmount,
			Provider:          testPaymentProvider,
			ProviderPaymentID: testProviderPaymentID,
			Currency:          testCurrency,
		},
	}
	wantJSON, err := json.Marshal(wantDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = suite.handler.HandleWebhookSuccess(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(wantJSON), rec.Body.String())
}

func (suite *paymentsHandlerTestSuite) TestHandleWebhookSuccess_EmptyPayload() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := suite.handler.HandleWebhookSuccess(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

type errorReader struct{}

var ErrRead = errors.New("read error")

func (e errorReader) Read(_ []byte) (int, error) {
	return 0, ErrRead
}

func (suite *paymentsHandlerTestSuite) TestHandleWebhookSuccess_InvalidPayload() {
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", errorReader{})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := suite.handler.HandleWebhookSuccess(c)
	suite.Require().Error(err)
	suite.Equal(http.StatusBadRequest, rec.Code)
}

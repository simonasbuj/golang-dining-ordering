package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"golang-dining-ordering/pkg/responses"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/services"
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
	testRestaurantID    = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRestaurantName  = "Test Restaurant"
	testCurrency        = "eur"
	testAmount          = 10
	testPaymentID       = uuid.MustParse("67676767-6767-4676-8767-676767676767")
	testOrderID         = uuid.MustParse("99999999-9999-4999-9999-999999999999")
	testDateTime        = time.Date(2025, time.December, 5, 19, 0, 0, 0, &time.Location{})
	testItemName        = "Test Menu Item"
	testOrderItemDto    = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	testItemID          = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	testCheckoutURL     = "http://fake-checkout-session.com/1"
	testPaymentProvider = db.OrdersPaymentProviderMock
)

type paymentsHandlerTestSuite struct {
	suite.Suite

	handler *PaymentsHandler
}

func (suite *paymentsHandlerTestSuite) SetupSuite() {
	mockOrdersRepo := NewMockOrdersRepo()
	mockPaymentsRepo := NewMockPaymentsRepo()
	mockPaymentsProvider := NewMockPaymentsProvider()
	svc := services.NewPaymentsService(mockOrdersRepo, mockPaymentsRepo, mockPaymentsProvider)

	suite.handler = NewPaymentsHandler(svc)
}

func TestPaymentsHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(paymentsHandlerTestSuite))
}

func (suite *paymentsHandlerTestSuite) TestHandleCreateCheckout() {
	e := echo.New()

	inputBodyDto := dto.CheckoutSessionRequestDto{
		OrderDto:   nil,
		SuccessURL: "https://fake-website.io?success=true",
		CancelURL:  "https://fake-website.io?cancel=true",
	}

	expectedResponseDto := responses.SuccessResponse{
		Message: "checkout session created",
		Data: &dto.CheckoutSessionResponseDto{
			URL:      testCheckoutURL,
			Provider: testPaymentProvider,
		},
	}
	expectedJSON, err := json.Marshal(expectedResponseDto)
	suite.Require().NoError(err)

	bodyBytes, err := json.Marshal(inputBodyDto)
	suite.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("order_id")
	c.SetParamValues(testOrderID.String())

	err = suite.handler.HandleCreateCheckout(c)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, rec.Code)
	suite.JSONEq(string(expectedJSON), rec.Body.String())
}

type mockPaymentsProvider struct {
	provider db.OrdersPaymentProvider
}

func NewMockPaymentsProvider() *mockPaymentsProvider {
	return &mockPaymentsProvider{
		provider: testPaymentProvider,
	}
}

func (p *mockPaymentsProvider) CreateCheckoutSession(
	_ context.Context,
	_ *dto.CheckoutSessionRequestDto,
) (*dto.CheckoutSessionResponseDto, error) {
	return &dto.CheckoutSessionResponseDto{
		URL:      testCheckoutURL,
		Provider: p.provider,
	}, nil
}

func (p *mockPaymentsProvider) VerifySuccessWebhookEvent(
	_ []byte,
	_ http.Header,
) (*dto.PaymentDto, error) {
	return &dto.PaymentDto{
		ID:                testPaymentID,
		OrderID:           testOrderID,
		AmountInCents:     10,
		Provider:          p.provider,
		ProviderPaymentID: "id",
		Currency:          "eur",
	}, nil
}

type mockPaymentsRepo struct{}

func NewMockPaymentsRepo() *mockPaymentsRepo {
	return &mockPaymentsRepo{}
}

func (r *mockPaymentsRepo) SavePayment(
	_ context.Context,
	_ *dto.PaymentDto,
) (*dto.PaymentDto, error) {
	return &dto.PaymentDto{
		ID:                testPaymentID,
		OrderID:           testOrderID,
		AmountInCents:     10,
		Provider:          db.OrdersPaymentProviderMock,
		ProviderPaymentID: "id",
		Currency:          "eur",
	}, nil
}

type mockOrdersRepo struct {
	orderDto *dto.OrderDto
}

func NewMockOrdersRepo() *mockOrdersRepo {
	return &mockOrdersRepo{
		orderDto: &dto.OrderDto{
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
					ID:           testOrderItemDto,
					RestaurantID: testRestaurantID,
					ItemID:       testItemID,
					Name:         testItemName,
					PriceInCents: testAmount,
				},
			},
		},
	}
}

func (r *mockOrdersRepo) GetCurrentOrderForTable(
	_ context.Context,
	_ uuid.UUID,
) (*dto.CurrentOrderDto, error) {
	return &dto.CurrentOrderDto{
		ID: testOrderID,
	}, nil
}

func (r *mockOrdersRepo) CreateOrderForTable(
	_ context.Context,
	_ uuid.UUID,
	_ string,
) (*dto.CurrentOrderDto, error) {
	return &dto.CurrentOrderDto{
		ID: testOrderID,
	}, nil
}

func (r *mockOrdersRepo) GetTableCurrency(_ context.Context, _ uuid.UUID) (string, error) {
	return testCurrency, nil
}

func (r *mockOrdersRepo) AddItemToOrder(
	_ context.Context,
	_ uuid.UUID,
	_ *dto.OrderItemDto,
) (uuid.UUID, error) {
	return testOrderItemDto, nil
}

func (r *mockOrdersRepo) GetOrderItems(
	_ context.Context,
	_ uuid.UUID,
) (*dto.OrderDto, error) {
	return r.orderDto, nil
}

func (r *mockOrdersRepo) GetMenuItem(
	_ context.Context,
	_ uuid.UUID,
) (*dto.OrderItemDto, error) {
	return r.orderDto.Items[0], nil
}

func (r *mockOrdersRepo) DeleteOrderItem(
	_ context.Context,
	_, _ uuid.UUID,
) error {
	return nil
}

func (r *mockOrdersRepo) UpdateOrder(_ context.Context, _ *dto.UpdateOrderReqDto) error {
	return nil
}

func (r *mockOrdersRepo) IsUserRestaurantWaiter(
	_ context.Context,
	_, _ uuid.UUID,
) error {
	return nil
}

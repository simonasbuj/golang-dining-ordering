package services

import (
	"context"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

//nolint:gochecknoglobals
var (
	testRestaurantID      = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRestaurantName    = "Test Restaurant"
	testCurrency          = "eur"
	testAmount            = 10
	testPaymentID         = uuid.MustParse("67676767-6767-4676-8767-676767676767")
	testOrderID           = uuid.MustParse("99999999-9999-4999-9999-999999999999")
	testDateTime          = time.Date(2025, time.December, 5, 19, 0, 0, 0, &time.Location{})
	testItemName          = "Test Menu Item"
	testOrderItemDto      = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	testItemID            = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	testCheckoutURL       = "http://fake-checkout-session.com/1"
	testPaymentProvider   = db.OrdersPaymentProviderMock
	testProviderPaymentID = "pi_123456"
)

type paymentsServiceTestSuite struct {
	suite.Suite

	svc *paymentsService
}

func (suite *paymentsServiceTestSuite) SetupSuite() {
	mockOrdersRepo := newMockOrdersRepo()
	mockPaymentsRepo := newMockPaymentsRepo()
	mockPaymentsProvider := newMockPaymentsProvider()
	suite.svc = NewPaymentsService(mockOrdersRepo, mockPaymentsRepo, mockPaymentsProvider)
}

func TestPaymentsServiceTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(paymentsServiceTestSuite))
}

func (suite *paymentsServiceTestSuite) TestCreateCheckout() {
	req := &dto.CheckoutSessionRequestDto{
		OrderDto:   nil,
		SuccessURL: "https://fake-url.com?success=true",
		CancelURL:  "https://fake-url.com?cancel=true",
	}

	want := &dto.CheckoutSessionResponseDto{
		URL:      testCheckoutURL,
		Provider: testPaymentProvider,
	}

	got, err := suite.svc.CreateCheckout(context.Background(), testOrderID, req)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *paymentsServiceTestSuite) TestHandleWebhookSuccess() {
	payload := []byte(`{"payment_secret": "secret"}`)
	header := http.Header{
		"Payment-Signature": []string{"signature"},
	}

	want := &dto.PaymentDto{
		ID:                testPaymentID,
		OrderID:           testOrderID,
		AmountInCents:     testAmount,
		Provider:          testPaymentProvider,
		ProviderPaymentID: testProviderPaymentID,
		Currency:          testCurrency,
	}

	got, err := suite.svc.HandleWebhookSuccess(context.Background(), payload, header)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *paymentsServiceTestSuite) TestCanPayForOrder_Status() {
	testCases := []struct {
		desc        string
		orderStatus db.OrderStatus
		wantCanPay  bool
		wantErr     error
	}{
		{"open order", db.OrderStatusOpen, true, nil},
		{"locked order", db.OrderStatusLocked, true, nil},
		{"canceled order", db.OrderStatusCancelled, false, ErrOrderFinalized},
		{"completed order", db.OrderStatusCompleted, false, ErrOrderFinalized},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.desc, func(_ *testing.T) {
			order := &dto.OrderDto{
				Status:            tc.orderStatus,
				TotalPriceInCents: 10,
			}
			got, err := suite.svc.canPayForOrder(order)

			if tc.wantErr != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.wantErr)
			} else {
				suite.Require().NoError(err)
			}

			suite.Require().Equal(tc.wantCanPay, got)
		})
	}
}

func (suite *paymentsServiceTestSuite) TestCanPayForOrder_Amount() {
	testCases := []struct {
		desc        string
		tipAmount   int
		totalAmount int
		wantCanPay  bool
		wantErr     error
	}{
		{"tip and total", 10, 10, true, nil},
		{"tip only", 10, 0, true, nil},
		{"total only", 0, 10, true, nil},
		{"no tip and total", 0, 0, true, ErrOrderPriceIsZero},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.desc, func(_ *testing.T) {
			order := &dto.OrderDto{
				Status:            db.OrderStatusOpen,
				TotalPriceInCents: tc.totalAmount,
				TipAmountInCents:  tc.tipAmount,
			}
			got, err := suite.svc.canPayForOrder(order)

			if tc.wantErr != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.wantErr)
			} else {
				suite.Require().NoError(err)
			}

			suite.Require().Equal(tc.wantCanPay, got)
		})
	}
}

type mockPaymentsProvider struct {
	provider db.OrdersPaymentProvider
}

func newMockPaymentsProvider() *mockPaymentsProvider {
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
		ProviderPaymentID: testProviderPaymentID,
		Currency:          testCurrency,
	}, nil
}

type mockPaymentsRepo struct{}

func newMockPaymentsRepo() *mockPaymentsRepo {
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
		ProviderPaymentID: testProviderPaymentID,
		Currency:          testCurrency,
	}, nil
}

type mockOrdersRepo struct {
	orderDto *dto.OrderDto
}

func newMockOrdersRepo() *mockOrdersRepo {
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

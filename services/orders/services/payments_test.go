package services

import (
	"context"
	"errors"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	mock "golang-dining-ordering/test/mock/orders"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

//nolint:gochecknoglobals
var (
	testRestaurantID              = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRestaurantName            = "Test Restaurant"
	testCurrency                  = "eur"
	testAmount                    = 10
	testPaymentID                 = uuid.MustParse("67676767-6767-4676-8767-676767676767")
	testOrderID                   = uuid.MustParse("99999999-9999-4999-9999-999999999999")
	testCompletedOrderID          = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	testTableID                   = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testDateTime                  = time.Date(2025, time.December, 5, 19, 0, 0, 0, &time.Location{})
	testItemName                  = "Test Menu Item"
	testOrderItemID               = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	testItemID                    = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	testDifferentRestaurantItemID = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-aaaaaaaaaaaa")
	testCheckoutURL               = "http://fake-checkout-session.com/1"
	testPaymentProvider           = db.OrdersPaymentProviderMock
	testProviderPaymentID         = "pi_123456"
)

var ErrPaymentProviderFailed = errors.New("payment provider failed")

type paymentsServiceTestSuite struct {
	suite.Suite

	svc *paymentsService
}

func (suite *paymentsServiceTestSuite) SetupSuite() {
	mockOrdersRepo := mock.NewMockOrdersRepo()
	mockPaymentsRepo := mock.NewMockPaymentsRepo()
	mockPaymentsProvider := mock.NewMockPaymentsProvider()
	suite.svc = NewPaymentsService(mockOrdersRepo, mockPaymentsRepo, mockPaymentsProvider)
}

func TestPaymentsServiceTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(paymentsServiceTestSuite))
}

func (suite *paymentsServiceTestSuite) TestCreateCheckout_Success() {
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

func (suite *paymentsServiceTestSuite) TestCreateCheckout_Error() {
	tests := []struct {
		name       string
		orderID    uuid.UUID
		successURL string
		cancelURL  string
	}{
		{"invalid order id", uuid.Max, "success.url", "cancel.url"},
		{"invalid dto", testOrderID, "", ""},
		{"order already paid", testCompletedOrderID, "success.url", "cancel.url"},
	}

	for _, tt := range tests {
		req := &dto.CheckoutSessionRequestDto{
			OrderDto:   nil,
			SuccessURL: tt.successURL,
			CancelURL:  tt.cancelURL,
		}

		got, err := suite.svc.CreateCheckout(context.Background(), tt.orderID, req)
		suite.Require().Error(err)
		suite.Nil(got)
	}
}

func (suite *paymentsServiceTestSuite) TestHandleWebhookSuccess_Success() {
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

func (suite *paymentsServiceTestSuite) TestHandleWebhookSuccess_Error() {
	tests := []struct {
		name       string
		ctxFailKey mock.CtxKey
		payload    []byte
	}{
		{"empty payload", "none", nil},
		{"save payment failed", "none", []byte("1")},
		{
			"repo failed to update order",
			mock.CtxFailUpdateOrder,
			[]byte(`{"payment_secret": "secret"}`),
		},
	}

	for _, tt := range tests {
		header := http.Header{
			"Payment-Signature": []string{"signature"},
		}

		ctx := context.WithValue(context.Background(), tt.ctxFailKey, true)
		got, err := suite.svc.HandleWebhookSuccess(ctx, tt.payload, header)
		suite.Require().Error(err)
		suite.Nil(got)
	}
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
		{"no tip and total", 0, 0, false, ErrOrderPriceIsZero},
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

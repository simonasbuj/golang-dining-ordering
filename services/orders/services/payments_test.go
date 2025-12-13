package services

import (
	"context"
	"errors"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/repository"
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

var (
	ErrRepoFailed            = errors.New("repository failed")
	ErrPaymentProviderFailed = errors.New("payment provider failed")
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

func (suite *paymentsServiceTestSuite) TestCreateCheckout_InvalidOrderID() {
	got, err := suite.svc.CreateCheckout(context.Background(), uuid.Max, nil)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *paymentsServiceTestSuite) TestCreateCheckout_InvalidReqDto() {
	req := &dto.CheckoutSessionRequestDto{
		OrderDto:   nil,
		SuccessURL: "",
		CancelURL:  "",
	}

	got, err := suite.svc.CreateCheckout(context.Background(), testOrderID, req)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *paymentsServiceTestSuite) TestCreateCheckout_OrderAlreadyPaid() {
	got, err := suite.svc.CreateCheckout(context.Background(), testCompletedOrderID, nil)
	suite.Require().Error(err)
	suite.Nil(got)
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

func (suite *paymentsServiceTestSuite) TestHandleWebhookSuccess_EmptyPayload() {
	got, err := suite.svc.HandleWebhookSuccess(context.Background(), nil, nil)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *paymentsServiceTestSuite) TestHandleWebhookSuccess_ErrorSavePayment() {
	payload := []byte("1")
	got, err := suite.svc.HandleWebhookSuccess(context.Background(), payload, nil)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *paymentsServiceTestSuite) TestHandleWebhookSuccess_ErrorUpdateOrder() {
	payload := []byte(`{"payment_secret": "secret"}`)
	header := http.Header{
		"Payment-Signature": []string{"signature"},
	}

	ctx := context.WithValue(context.Background(), ctxFailUpdateOrder, true)
	got, err := suite.svc.HandleWebhookSuccess(ctx, payload, header)
	suite.Require().Error(err)
	suite.Nil(got)
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
	req *dto.CheckoutSessionRequestDto,
) (*dto.CheckoutSessionResponseDto, error) {
	if req.SuccessURL == "" {
		return nil, ErrPaymentProviderFailed
	}

	return &dto.CheckoutSessionResponseDto{
		URL:      testCheckoutURL,
		Provider: p.provider,
	}, nil
}

func (p *mockPaymentsProvider) VerifySuccessWebhookEvent(
	payload []byte,
	_ http.Header,
) (*dto.PaymentDto, error) {
	if len(payload) == 0 {
		return nil, ErrPaymentProviderFailed
	}

	if len(payload) == 1 {
		return &dto.PaymentDto{}, nil
	}

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
	reqDto *dto.PaymentDto,
) (*dto.PaymentDto, error) {
	if reqDto.OrderID == uuid.Nil {
		return nil, ErrRepoFailed
	}

	return &dto.PaymentDto{
		ID:                testPaymentID,
		OrderID:           testOrderID,
		AmountInCents:     10,
		Provider:          db.OrdersPaymentProviderMock,
		ProviderPaymentID: testProviderPaymentID,
		Currency:          testCurrency,
	}, nil
}

type ctxKey string

const (
	ctxFailUpdateOrder         ctxKey = "fail-UpdateOrder"
	ctxFailGetTableCurrency    ctxKey = "fail-GetTableCurrency"
	ctxFailCreateOrderForTable ctxKey = "fail-CreateOrderForTable"
	ctxFailAddItemToOrder      ctxKey = "fail-AddItemToOrder"
)

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
					ID:           testOrderItemID,
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
	tableID uuid.UUID,
) (*dto.CurrentOrderDto, error) {
	if tableID == uuid.Max {
		return nil, ErrRepoFailed
	}

	if tableID != testTableID {
		return nil, repository.ErrNoCurrentOrder
	}

	return &dto.CurrentOrderDto{
		ID: testOrderID,
	}, nil
}

func (r *mockOrdersRepo) CreateOrderForTable(
	ctx context.Context,
	_ uuid.UUID,
	_ string,
) (*dto.CurrentOrderDto, error) {
	if v, ok := ctx.Value(ctxFailCreateOrderForTable).(bool); ok && v {
		return nil, ErrRepoFailed
	}

	return &dto.CurrentOrderDto{
		ID: testOrderID,
	}, nil
}

func (r *mockOrdersRepo) GetTableCurrency(ctx context.Context, _ uuid.UUID) (string, error) {
	if v, ok := ctx.Value(ctxFailGetTableCurrency).(bool); ok && v {
		return "", ErrRepoFailed
	}

	return testCurrency, nil
}

func (r *mockOrdersRepo) AddItemToOrder(
	ctx context.Context,
	_ uuid.UUID,
	_ *dto.OrderItemDto,
) (*dto.OrderItemDto, error) {
	if v, ok := ctx.Value(ctxFailAddItemToOrder).(bool); ok && v {
		return nil, ErrRepoFailed
	}

	orderItemDto := &dto.OrderItemDto{
		ID:           testOrderItemID,
		RestaurantID: testRestaurantID,
		ItemID:       testItemID,
		Name:         testItemName,
		PriceInCents: testAmount,
	}

	return orderItemDto, nil
}

func (r *mockOrdersRepo) GetOrderItems(
	_ context.Context,
	orderID uuid.UUID,
) (*dto.OrderDto, error) {
	if orderID == testCompletedOrderID {
		completedOrder := *r.orderDto
		completedOrder.Status = db.OrderStatusCompleted

		return &completedOrder, nil
	}

	if orderID != testOrderID {
		return nil, ErrRepoFailed
	}

	respDto := *r.orderDto

	return &respDto, nil
}

func (r *mockOrdersRepo) GetMenuItem(
	_ context.Context,
	itemID uuid.UUID,
) (*dto.OrderItemDto, error) {
	if itemID == testDifferentRestaurantItemID {
		item := *r.orderDto.Items[0]
		item.RestaurantID = uuid.Max

		return &item, nil
	}

	if itemID != testItemID {
		return nil, ErrRepoFailed
	}

	return r.orderDto.Items[0], nil
}

func (r *mockOrdersRepo) DeleteOrderItem(
	_ context.Context,
	orderItemID, _ uuid.UUID,
) (*dto.OrderItemDto, error) {
	if orderItemID != testOrderItemID {
		return nil, ErrRepoFailed
	}

	return &dto.OrderItemDto{
		ID:           testOrderItemID,
		RestaurantID: testRestaurantID,
		ItemID:       testItemID,
		Name:         testItemName,
		PriceInCents: testAmount,
	}, nil
}

func (r *mockOrdersRepo) UpdateOrder(
	ctx context.Context,
	req *dto.UpdateOrderReqDto,
) (*dto.OrderDto, error) {
	if v, ok := ctx.Value(ctxFailUpdateOrder).(bool); ok && v {
		return nil, ErrRepoFailed
	}

	status := db.OrderStatusOpen
	if req.Status != nil {
		status = *req.Status
	}

	tip := testAmount
	if req.TipAmountInCents != nil {
		tip = int(*req.TipAmountInCents)
	}

	return &dto.OrderDto{
		ID:               req.OrderID,
		Status:           status,
		TipAmountInCents: tip,
	}, nil
}

func (r *mockOrdersRepo) IsUserRestaurantWaiter(
	_ context.Context,
	userID, _ uuid.UUID,
) error {
	if userID == testUserFromAnotherRestaurantID {
		return ErrRepoFailed
	}

	return nil
}

func (r *mockOrdersRepo) AssignWaiter(_ context.Context, orderID, _ uuid.UUID) error {
	if orderID != testOrderID {
		return ErrRepoFailed
	}

	return nil
}

func (r *mockOrdersRepo) RemoveWaiter(_ context.Context, orderID, _, _ uuid.UUID) error {
	if orderID != testOrderID {
		return ErrRepoFailed
	}

	return nil
}

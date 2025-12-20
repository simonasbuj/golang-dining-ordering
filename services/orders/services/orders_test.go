package services

import (
	"context"
	authDto "golang-dining-ordering/services/auth/dto"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	mock "golang-dining-ordering/test/mock/orders"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

//nolint:gochecknoglobals
var (
	testUserID                      = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	testUserFromAnotherRestaurantID = uuid.MustParse("69696969-6969-6969-6969-696969696969")
)

type ordersServiceTestSuite struct {
	suite.Suite

	svc      *ordersService
	orderDto *dto.OrderDto
}

func (suite *ordersServiceTestSuite) SetupSuite() {
	mockOrdersRepo := mock.NewMockOrdersRepo()
	suite.svc = NewOrdersService(mockOrdersRepo)

	suite.orderDto = &dto.OrderDto{
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

func TestOrdersServiceTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ordersServiceTestSuite))
}

func (suite *ordersServiceTestSuite) TestGetOrder_Success() {
	want := *suite.orderDto

	got, err := suite.svc.GetOrder(context.Background(), testOrderID)
	suite.Require().NoError(err)
	suite.Equal(&want, got)
}

func (suite *ordersServiceTestSuite) TestGetOrder_RepoFailed() {
	got, err := suite.svc.GetOrder(context.Background(), uuid.Max)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestGetOrCreateCurrentOrderForTable_SuccessOrderExists() {
	want := &dto.CurrentOrderDto{
		ID: testOrderID,
	}

	got, err := suite.svc.GetOrCreateCurrentOrderForTable(context.Background(), testTableID)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *ordersServiceTestSuite) TestGetOrCreateCurrentOrderForTable_SuccessNewOrderCreated() {
	want := &dto.CurrentOrderDto{
		ID: testOrderID,
	}

	got, err := suite.svc.GetOrCreateCurrentOrderForTable(
		context.Background(),
		testCompletedOrderID,
	)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *ordersServiceTestSuite) TestGetOrCreateCurrentOrderForTable_Error() {
	tests := []struct {
		name       string
		failCtxKey mock.CtxKey
		orderID    uuid.UUID
	}{
		{"repo error", "none", uuid.Max},
		{"repo failed getting table currency", mock.CtxFailGetTableCurrency, testOrderID},
		{"repo failed creating new order", mock.CtxFailCreateOrderForTable, testOrderID},
	}

	for _, tt := range tests {
		ctx := context.WithValue(context.Background(), tt.failCtxKey, true)
		got, err := suite.svc.GetOrCreateCurrentOrderForTable(ctx, tt.orderID)
		suite.Require().Error(err)
		suite.Nil(got)
	}
}

func (suite *ordersServiceTestSuite) TestAddItemToOrder_Success() {
	want := *suite.orderDto
	want.Items = append(want.Items, &dto.OrderItemDto{
		ID:           testOrderItemID,
		RestaurantID: testRestaurantID,
		ItemID:       testItemID,
		Name:         testItemName,
		PriceInCents: testAmount,
	})
	want.TotalPriceInCents += 10

	got, err := suite.svc.AddItemToOrder(context.Background(), testOrderID, testItemID)
	suite.Require().NoError(err)
	suite.Equal(&want, got)
}

func (suite *ordersServiceTestSuite) TestAddItemToOrder_Error() {
	tests := []struct {
		name       string
		failCtxKey mock.CtxKey
		orderID    uuid.UUID
		itemID     uuid.UUID
	}{
		{"repo failed get menu items", "none", testOrderID, uuid.Max},
		{"repo failed get order items", "none", uuid.Max, testItemID},
		{"cant add complete order", "none", testCompletedOrderID, testItemID},
		{"adding from another restaurant", "none", testOrderID, testDifferentRestaurantItemID},
		{"repo failed adding new order", mock.CtxFailAddItemToOrder, testOrderID, testItemID},
	}

	for _, tt := range tests {
		ctx := context.WithValue(context.Background(), tt.failCtxKey, true)
		got, err := suite.svc.AddItemToOrder(ctx, tt.orderID, tt.itemID)
		suite.Require().Error(err)
		suite.Nil(got)
	}
}

func (suite *ordersServiceTestSuite) TestDeleteOrderItem_Success() {
	want := *suite.orderDto
	want.Items = []*dto.OrderItemDto{}
	want.TotalPriceInCents = 0
	got, err := suite.svc.DeleteOrderItem(context.Background(), testOrderItemID, testOrderID)
	suite.Require().NoError(err)
	suite.Equal(&want, got)
}

func (suite *ordersServiceTestSuite) TestDeleteOrderItem_Error() {
	tests := []struct {
		name        string
		orderItemID uuid.UUID
		orderID     uuid.UUID
	}{
		{"repo failed get order items", testOrderItemID, uuid.Max},
		{"cant delete from locked order", testOrderItemID, testCompletedOrderID},
		{"repo failed delete order item", uuid.Max, testOrderID},
	}

	for _, tt := range tests {
		got, err := suite.svc.DeleteOrderItem(context.Background(), tt.orderItemID, tt.orderID)
		suite.Require().Error(err)
		suite.Nil(got)
	}
}

func (suite *ordersServiceTestSuite) TestUpdateOrder_Success() {
	status := db.OrderStatusLocked
	tip := int32(testAmount) //nolint:gosec
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: &tip,
		Status:           &status,
	}

	want := *suite.orderDto
	want.Status = status
	got, err := suite.svc.UpdateOrder(context.Background(), reqDto, nil)
	suite.Require().NoError(err)
	suite.Equal(&want, got)
}

func (suite *ordersServiceTestSuite) TestUpdateOrder_Error() {
	statusLocked := db.OrderStatusLocked
	tip := int32(testAmount) //nolint:gosec
	tests := []struct {
		name             string
		ctxFailKey       mock.CtxKey
		orderID          uuid.UUID
		tipAmountInCents *int32
		status           *db.OrderStatus
	}{
		{"empty payload", "none", testOrderID, nil, nil},
		{"repo failed get order items", "none", uuid.Max, &tip, &statusLocked},
		{"cant edit locked order", "none", testCompletedOrderID, &tip, &statusLocked},
		{"repo failed update order", mock.CtxFailUpdateOrder, testOrderID, &tip, &statusLocked},
	}

	for _, tt := range tests {
		reqDto := &dto.UpdateOrderReqDto{
			OrderID:          tt.orderID,
			TipAmountInCents: tt.tipAmountInCents,
			Status:           tt.status,
		}

		ctx := context.WithValue(context.Background(), tt.ctxFailKey, true)
		got, err := suite.svc.UpdateOrder(ctx, reqDto, nil)
		suite.Require().Error(err)
		suite.Nil(got)
	}
}

func (suite *ordersServiceTestSuite) TestCanUserEditOrder_Success() {
	orderDto := &dto.OrderDto{
		Status: db.OrderStatusOpen,
	}

	status := db.OrderStatusLocked
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: nil,
		Status:           &status,
	}

	got, err := suite.svc.canUserEditOrder(context.Background(), orderDto, nil, reqDto)
	suite.Require().NoError(err)
	suite.True(got)
}

func (suite *ordersServiceTestSuite) TestCanUserEditOrder_Error() {
	statusLocked := db.OrderStatusLocked
	statusCanceled := db.OrderStatusCancelled

	tests := []struct {
		name      string
		status    *db.OrderStatus
		userID    uuid.UUID
		wantError error
	}{
		{"customer cant edit locked order", &statusLocked, uuid.Nil, ErrUserCannotEditLockedOrder},
		{"customer cant set order to canceled", &statusCanceled, uuid.Nil, ErrUserCannotEditStatus},
		{
			"waiter from another restaurant cant edit order",
			&statusCanceled,
			testUserFromAnotherRestaurantID,
			ErrUserCannotEditLockedOrder,
		},
	}

	for _, tt := range tests {
		orderDto := &dto.OrderDto{
			Status: db.OrderStatusLocked,
		}

		reqDto := &dto.UpdateOrderReqDto{
			OrderID:          testOrderID,
			TipAmountInCents: nil,
			Status:           tt.status,
		}

		claims := &authDto.TokenClaimsDto{
			UserID: tt.userID,
		}

		got, err := suite.svc.canUserEditOrder(context.Background(), orderDto, claims, reqDto)
		suite.Require().Error(err)
		suite.Require().ErrorIs(err, tt.wantError)
		suite.False(got)
	}
}

func (suite *ordersServiceTestSuite) TestAssignWaiter_Success() {
	err := suite.svc.AssignWaiter(context.Background(), testOrderID, testUserID)
	suite.Require().NoError(err)
}

func (suite *ordersServiceTestSuite) TestAssignWaiter_Error() {
	tests := []struct {
		name    string
		orderID uuid.UUID
		userID  uuid.UUID
	}{
		{"failed to get current order", uuid.Nil, testUserID},
		{"user is not restaurant waiter", testOrderID, testUserFromAnotherRestaurantID},
		{"repo failed", testCompletedOrderID, testUserID},
	}

	for _, tt := range tests {
		err := suite.svc.AssignWaiter(context.Background(), tt.orderID, tt.userID)
		suite.Require().Error(err)
	}
}

func (suite *ordersServiceTestSuite) TestRemoveWaiter_Success() {
	err := suite.svc.RemoveWaiter(context.Background(), testOrderID, testUserID, uuid.New())
	suite.Require().NoError(err)
}

func (suite *ordersServiceTestSuite) TestRemoveWaiter_Error() {
	tests := []struct {
		name    string
		orderID uuid.UUID
		userID  uuid.UUID
	}{
		{"failed to get current order", uuid.Nil, testUserID},
		{"user is not restaurant waiter", testOrderID, testUserFromAnotherRestaurantID},
		{"repo failed", testCompletedOrderID, testUserID},
	}

	for _, tt := range tests {
		err := suite.svc.RemoveWaiter(context.Background(), tt.orderID, tt.userID, uuid.New())
		suite.Require().Error(err)
	}
}

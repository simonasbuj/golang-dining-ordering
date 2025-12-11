package services

import (
	"context"
	authDto "golang-dining-ordering/services/auth/dto"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

//nolint:gochecknoglobals
var (
	testUserFromAnotherRestaurantID = uuid.MustParse("69696969-6969-6969-6969-696969696969")
)

type ordersServiceTestSuite struct {
	suite.Suite

	svc      *ordersService
	orderDto *dto.OrderDto
}

func (suite *ordersServiceTestSuite) SetupSuite() {
	mockOrdersRepo := newMockOrdersRepo()
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

func (suite *ordersServiceTestSuite) TestGetOrCreateCurrentOrderForTable_RepoError() {
	got, err := suite.svc.GetOrCreateCurrentOrderForTable(context.Background(), uuid.Max)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestGetOrCreateCurrentOrderForTable_FailedGettingTableCurrency() {
	ctx := context.WithValue(context.Background(), ctxFailGetTableCurrency, true)
	got, err := suite.svc.GetOrCreateCurrentOrderForTable(ctx, testOrderID)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestGetOrCreateCurrentOrderForTable_FailedCreatingNewOrderForTable() {
	ctx := context.WithValue(context.Background(), ctxFailCreateOrderForTable, true)
	got, err := suite.svc.GetOrCreateCurrentOrderForTable(ctx, testOrderID)
	suite.Require().Error(err)
	suite.Nil(got)
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

func (suite *ordersServiceTestSuite) TestAddItemToOrder_FailedGetMenuItems() {
	got, err := suite.svc.AddItemToOrder(context.Background(), testOrderID, uuid.Max)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestAddItemToOrder_FailedGetOrderItems() {
	got, err := suite.svc.AddItemToOrder(context.Background(), uuid.Max, testItemID)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestAddItemToOrder_CantAddToCompletedOrder() {
	got, err := suite.svc.AddItemToOrder(context.Background(), testCompletedOrderID, testItemID)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrOrderIsNotOpen)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestAddItemToOrder_TryingToAddItemFromAnotherRestaurant() {
	got, err := suite.svc.AddItemToOrder(
		context.Background(),
		testOrderID,
		testDifferentRestaurantItemID,
	)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrItemDoesNotBelongToRestaurant)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestAddItemToOrder_FailedRepoAddingItemToOrder() {
	ctx := context.WithValue(context.Background(), ctxFailAddItemToOrder, true)
	got, err := suite.svc.AddItemToOrder(ctx, testOrderID, testItemID)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestDeleteOrderItem_Success() {
	want := *suite.orderDto
	want.Items = []*dto.OrderItemDto{}
	want.TotalPriceInCents = 0
	got, err := suite.svc.DeleteOrderItem(context.Background(), testOrderItemID, testOrderID)
	suite.Require().NoError(err)
	suite.Equal(&want, got)
}

func (suite *ordersServiceTestSuite) TestDeleteOrderItem_FailedGetOrderItems() {
	got, err := suite.svc.DeleteOrderItem(context.Background(), testOrderItemID, uuid.Max)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestDeleteOrderItem_CantDeleteFromLockedOrder() {
	got, err := suite.svc.DeleteOrderItem(
		context.Background(),
		testOrderItemID,
		testCompletedOrderID,
	)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrOrderIsNotOpen)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestDeleteOrderItem_FailedRepoDeleteOrderItem() {
	got, err := suite.svc.DeleteOrderItem(context.Background(), uuid.Max, testOrderID)
	suite.Require().Error(err)
	suite.Nil(got)
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

func (suite *ordersServiceTestSuite) TestUpdateOrder_EmptyPayload() {
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: nil,
		Status:           nil,
	}

	got, err := suite.svc.UpdateOrder(context.Background(), reqDto, nil)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrPayloadEmpty)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestUpdateOrder_RepoFailedGetOrderItems() {
	status := db.OrderStatusLocked
	tip := int32(testAmount) //nolint:gosec
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          uuid.Max,
		TipAmountInCents: &tip,
		Status:           &status,
	}

	got, err := suite.svc.UpdateOrder(context.Background(), reqDto, nil)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestUpdateOrder_CantEditCompletedOrder() {
	status := db.OrderStatusLocked
	tip := int32(testAmount) //nolint:gosec
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testCompletedOrderID,
		TipAmountInCents: &tip,
		Status:           &status,
	}

	got, err := suite.svc.UpdateOrder(context.Background(), reqDto, nil)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrOrderFinalized)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestUpdateOrder_RepoFailedUpdateOrder() {
	status := db.OrderStatusLocked
	tip := int32(testAmount) //nolint:gosec
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: &tip,
		Status:           &status,
	}

	ctx := context.WithValue(context.Background(), ctxFailUpdateOrder, true)
	got, err := suite.svc.UpdateOrder(ctx, reqDto, nil)
	suite.Require().Error(err)
	suite.Nil(got)
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

func (suite *ordersServiceTestSuite) TestCanUserEditOrder_UserCannotEditLockedOrder() {
	orderDto := &dto.OrderDto{
		Status: db.OrderStatusLocked,
	}

	status := db.OrderStatusLocked
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: nil,
		Status:           &status,
	}

	claims := &authDto.TokenClaimsDto{
		UserID: uuid.Nil,
	}

	got, err := suite.svc.canUserEditOrder(context.Background(), orderDto, claims, reqDto)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrUserCannotEditLockedOrder)
	suite.False(got)
}

func (suite *ordersServiceTestSuite) TestCanUserEditOrder_UserCannotEditStatus() {
	orderDto := &dto.OrderDto{
		Status: db.OrderStatusLocked,
	}

	status := db.OrderStatusCancelled
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: nil,
		Status:           &status,
	}

	claims := &authDto.TokenClaimsDto{
		UserID: uuid.Nil,
	}

	got, err := suite.svc.canUserEditOrder(context.Background(), orderDto, claims, reqDto)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrUserCannotEditStatus)
	suite.False(got)
}

func (suite *ordersServiceTestSuite) TestCanUserEditOrder_WaiterCantEditOrder() {
	orderDto := &dto.OrderDto{
		Status: db.OrderStatusLocked,
	}

	status := db.OrderStatusCancelled
	reqDto := &dto.UpdateOrderReqDto{
		OrderID:          testOrderID,
		TipAmountInCents: nil,
		Status:           &status,
	}

	claims := &authDto.TokenClaimsDto{
		UserID: testUserFromAnotherRestaurantID,
	}

	got, err := suite.svc.canUserEditOrder(context.Background(), orderDto, claims, reqDto)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrUserCannotEditLockedOrder)
	suite.False(got)
}

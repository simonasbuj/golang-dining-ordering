package services

import (
	"context"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
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
				ID:           testOrderItemDto,
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
	want := suite.orderDto

	got, err := suite.svc.GetOrder(context.Background(), testOrderID)
	suite.Require().NoError(err)
	suite.Equal(want, got)
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
	want := suite.orderDto

	got, err := suite.svc.AddItemToOrder(context.Background(), testOrderID, testItemID)
	suite.Require().NoError(err)
	suite.Equal(want, got)
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
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestAddItemToOrder_TryingToAddItemFromAnotherRestaurant() {
	got, err := suite.svc.AddItemToOrder(
		context.Background(),
		testOrderID,
		testDifferentRestaurantItemID,
	)
	suite.Require().Error(err)
	suite.Nil(got)
}

func (suite *ordersServiceTestSuite) TestAddItemToOrder_FailedRepoAddingItemToOrder() {
	ctx := context.WithValue(context.Background(), ctxFailAddItemToOrder, true)
	got, err := suite.svc.AddItemToOrder(ctx, testOrderID, testItemID)
	suite.Require().Error(err)
	suite.Nil(got)
}

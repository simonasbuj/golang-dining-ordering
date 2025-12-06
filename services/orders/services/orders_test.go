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

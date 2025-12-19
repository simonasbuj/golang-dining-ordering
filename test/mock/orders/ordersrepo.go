// Package orders holds implementation of moclk orders/payments repos and mock payment provider.
package orders

import (
	"context"
	"errors"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"golang-dining-ordering/services/orders/repository"
	"time"

	"github.com/google/uuid"
)

//nolint:gochecknoglobals
var (
	testRestaurantID     = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testRestaurantName   = "Test Restaurant"
	testCurrency         = "eur"
	testAmount           = 10
	testPaymentID        = uuid.MustParse("67676767-6767-4676-8767-676767676767")
	testOrderID          = uuid.MustParse("99999999-9999-4999-9999-999999999999")
	testCompletedOrderID = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	testTableID          = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	testDateTime         = time.Date(
		2025,
		time.December,
		5,
		19,
		0,
		0,
		0,
		&time.Location{},
	)
	testItemName                    = "Test Menu Item"
	testOrderItemID                 = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	testItemID                      = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	testDifferentRestaurantItemID   = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-aaaaaaaaaaaa")
	testCheckoutURL                 = "http://fake-checkout-session.com/1"
	testPaymentProvider             = db.OrdersPaymentProviderMock
	testProviderPaymentID           = "pi_123456"
	testUserFromAnotherRestaurantID = uuid.MustParse("69696969-6969-6969-6969-696969696969")
)

var (
	// ErrRepoFailed is returned when mock repo fails.
	ErrRepoFailed = errors.New("repository failed")
	// ErrPaymentProviderFailed is returned when mock payment provider fails.
	ErrPaymentProviderFailed = errors.New("payment provider failed")
)

type ctxKey string

const (
	// CtxFailUpdateOrder is a context key to simulate UpdateOrder failure in tests.
	CtxFailUpdateOrder ctxKey = "fail-UpdateOrder"
	// CtxFailGetTableCurrency is a context key to simulate GetTableCurrency failure in tests.
	CtxFailGetTableCurrency ctxKey = "fail-GetTableCurrency"
	// CtxFailCreateOrderForTable is a context key to simulate CreateOrderForTable failure in tests.
	CtxFailCreateOrderForTable ctxKey = "fail-CreateOrderForTable"
	// CtxFailAddItemToOrder is a context key to simulate AddItemToOrder failure in tests.
	CtxFailAddItemToOrder ctxKey = "fail-AddItemToOrder"
)

type mockOrdersRepo struct {
	orderDto *dto.OrderDto
}

func NewMockOrdersRepo() *mockOrdersRepo { //nolint:revive
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
	if v, ok := ctx.Value(CtxFailCreateOrderForTable).(bool); ok && v {
		return nil, ErrRepoFailed
	}

	return &dto.CurrentOrderDto{
		ID: testOrderID,
	}, nil
}

func (r *mockOrdersRepo) GetTableCurrency(ctx context.Context, _ uuid.UUID) (string, error) {
	if v, ok := ctx.Value(CtxFailGetTableCurrency).(bool); ok && v {
		return "", ErrRepoFailed
	}

	return testCurrency, nil
}

func (r *mockOrdersRepo) AddItemToOrder(
	ctx context.Context,
	_ uuid.UUID,
	_ *dto.OrderItemDto,
) (*dto.OrderItemDto, error) {
	if v, ok := ctx.Value(CtxFailAddItemToOrder).(bool); ok && v {
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
	if v, ok := ctx.Value(CtxFailUpdateOrder).(bool); ok && v {
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

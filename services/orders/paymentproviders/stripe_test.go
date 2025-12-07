package paymentproviders

import (
	"golang-dining-ordering/services/orders/dto"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:gochecknoglobals
var (
	testCurrency   = "eur"
	testTipAmount  = 2000
	testItem1Name  = "Žalgirio Sumuštinis"
	testItem1Price = 5000
	testItem2Name  = "Bananinis Miau"
	testItem2Price = 4000
)

func TestCreateLineItems(t *testing.T) {
	t.Parallel()

	provider := &StripePaymentProvider{}

	order := &dto.OrderDto{
		Currency:         testCurrency,
		TipAmountInCents: testTipAmount,
		Items: []*dto.OrderItemDto{
			{Name: testItem1Name, PriceInCents: testItem1Price},
			{Name: testItem2Name, PriceInCents: testItem2Price},
		},
	}

	lineItems := provider.createLineItems(order)
	assert.Len(t, lineItems, 3)

	assert.Equal(t, testItem1Name, *lineItems[0].PriceData.ProductData.Name)
	assert.Equal(t, int64(testItem1Price), *lineItems[0].PriceData.UnitAmount)
	assert.Equal(t, testCurrency, *lineItems[0].PriceData.Currency)

	assert.Equal(t, testItem2Name, *lineItems[1].PriceData.ProductData.Name)
	assert.Equal(t, int64(testItem2Price), *lineItems[1].PriceData.UnitAmount)

	tip := lineItems[2]
	assert.Equal(t, "Tip for the staff", *tip.PriceData.ProductData.Name)
	assert.Equal(t, int64(testTipAmount), *tip.PriceData.UnitAmount)
	assert.Equal(t, testCurrency, *tip.PriceData.Currency)

	for i, li := range lineItems {
		assert.Equal(t, int64(1), *li.Quantity, "line item %d quantity", i)
	}
}

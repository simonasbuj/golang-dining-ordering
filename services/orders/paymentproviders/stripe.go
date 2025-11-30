package paymentproviders

import (
	"context"
	"fmt"
	"golang-dining-ordering/services/orders/dto"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
)

// StripePaymentProvider implements the PaymentProvider interface.
type StripePaymentProvider struct{}

// NewStripePaymentProvider creates an instance of stripe payment provider
// it panics if secretKey is not provided.
func NewStripePaymentProvider(secretKey string) *StripePaymentProvider {
	if secretKey == "" {
		panic("secretKey is required for StripePaymentProvider")
	}

	stripe.Key = secretKey

	return &StripePaymentProvider{}
}

// CreateCheckoutSession creates checkout session for provided order items and returns it's url.
func (p *StripePaymentProvider) CreateCheckoutSession(
	_ context.Context,
	reqDto *dto.CheckoutSessionRequestDto,
) (string, error) {
	order := reqDto.OrderDto

	lineItems := p.createLineItems(order)

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(reqDto.SuccessURL),
		CancelURL:  stripe.String(reqDto.CancelURL),
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"order_id": reqDto.OrderDto.ID.String(),
			},
		},
		LineItems: lineItems,
	}

	s, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("creating stripe checkout session: %w", err)
	}

	return s.URL, nil
}

func (p *StripePaymentProvider) createLineItems(
	order *dto.OrderDto,
) []*stripe.CheckoutSessionLineItemParams {
	lineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(order.Items)+1)

	for _, item := range order.Items {
		lineItem := &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(order.Currency),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(item.Name),
				},
				UnitAmount: stripe.Int64(int64(item.PriceInCents)),
			},
			Quantity: stripe.Int64(1),
		}

		lineItems = append(lineItems, lineItem)
	}

	tipLineItem := &stripe.CheckoutSessionLineItemParams{
		PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency: stripe.String(order.Currency),
			ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
				Name: stripe.String("Tip for the staff"),
			},
			UnitAmount: stripe.Int64(int64(order.TipAmountInCents)),
		},
		Quantity: stripe.Int64(1),
	}

	lineItems = append(lineItems, tipLineItem)

	return lineItems
}

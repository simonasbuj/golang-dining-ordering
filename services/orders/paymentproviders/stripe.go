package paymentproviders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	db "golang-dining-ordering/services/orders/db/generated"
	"golang-dining-ordering/services/orders/dto"
	"net/http"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"
)

var (
	// ErrUnknownWebhookEventType is returned when unknown Stripe webhook event is received.
	ErrUnknownWebhookEventType = errors.New("unhandled event type")
	// ErrOrderIDMissingInMetadata is returned when 'order_id' is missing from Stripe payment intent metadata.
	ErrOrderIDMissingInMetadata = errors.New("order_id missing from payment metadata")
)

const metadataKeyOrderID = "order_id"

// StripePaymentProvider implements the PaymentProvider interface.
type StripePaymentProvider struct {
	webhookSecret string
}

// NewStripePaymentProvider creates an instance of stripe payment provider
// it panics if secretKey is not provided.
func NewStripePaymentProvider(secretKey, webhookSecret string) *StripePaymentProvider {
	if secretKey == "" || webhookSecret == "" {
		panic("secretKey and webhookSecret is required for StripePaymentProvider")
	}

	stripe.Key = secretKey

	return &StripePaymentProvider{
		webhookSecret: webhookSecret,
	}
}

// CreateCheckoutSession creates checkout session for provided order items and returns it's url.
func (p *StripePaymentProvider) CreateCheckoutSession(
	_ context.Context,
	reqDto *dto.CheckoutSessionRequestDto,
) (*dto.CheckoutSessionResponseDto, error) {
	order := reqDto.OrderDto

	lineItems := p.createLineItems(order)

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(reqDto.SuccessURL),
		CancelURL:  stripe.String(reqDto.CancelURL),
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				metadataKeyOrderID: reqDto.OrderDto.ID.String(),
			},
		},
		LineItems: lineItems,
	}

	s, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("creating stripe checkout session: %w", err)
	}

	respDto := &dto.CheckoutSessionResponseDto{
		URL:      s.URL,
		Provider: db.OrdersPaymentProviderStripe,
	}

	return respDto, nil
}

// VerifySuccessWebhookEvent handles successful payment webhook request.
func (p *StripePaymentProvider) VerifySuccessWebhookEvent(
	payload []byte,
	header http.Header,
) (*dto.PaymentDto, error) {
	sigHeader := header.Get("Stripe-Signature")

	event, err := webhook.ConstructEvent(payload, sigHeader, p.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("veryfing stripe webhook signature: %w", err)
	}

	if event.Type != "payment_intent.succeeded" {
		return nil, fmt.Errorf("%w: %s", ErrUnknownWebhookEventType, event.Type)
	}

	var pi stripe.PaymentIntent

	err = json.Unmarshal(event.Data.Raw, &pi)
	if err != nil {
		return nil, fmt.Errorf("unmarhsaling payment intent: %w", err)
	}

	orderIDstr, ok := pi.Metadata[metadataKeyOrderID]
	if !ok || orderIDstr == "" {
		return nil, ErrOrderIDMissingInMetadata
	}

	orderID, err := uuid.Parse(orderIDstr)
	if err != nil {
		return nil, fmt.Errorf("parsing orderID from payment intent: %w", err)
	}

	respDto := &dto.PaymentDto{
		ID:                uuid.New(),
		OrderID:           orderID,
		AmountInCents:     int(pi.AmountReceived),
		Provider:          db.OrdersPaymentProviderStripe,
		ProviderPaymentID: pi.ID,
		Currency:          string(pi.Currency),
	}

	return respDto, nil
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

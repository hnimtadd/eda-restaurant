package stripeengine

import (
	"log"
	"os"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

type StripeEngine struct {
}

func NewStripeEngine() StripeEngine {
	engine := StripeEngine{}
	if err := engine.initStrip(); err != nil {
		log.Fatal(err)
	}
	return engine
}
func (s *StripeEngine) initStrip() error {
	stripe.Key = os.Getenv("stripeKey")
	return nil
}
func (s *StripeEngine) CreateCheckOutSession(money float64) (*stripe.CheckoutSession, error) {
	log.Println("making checkout session", money)
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:          stripe.String(string(stripe.CurrencyVND)),
					UnitAmountDecimal: stripe.Float64(money),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("orderProduct"),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String("https://example.com/success"),
		CancelURL:  stripe.String("https://example.com/cancel"),
		ExpiresAt:  stripe.Int64(time.Now().Add(30 * time.Minute).Unix()),
	}

	checkoutSession, err := session.New(params)

	if err != nil {
		log.Printf("[Stripe Engine]session.New: %v", err)
		return nil, err
	}
	return checkoutSession, nil
}

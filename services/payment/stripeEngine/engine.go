package stripeengine

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	key := viper.GetViper().GetString("stripe_key")
	if key == "" {
		log.Debugf("stripe_key must specified in .env file")
		return errors.New("stripe key must not null")
	}
	return nil
}
func (s *StripeEngine) CreateCheckOutSession(money float64) (*stripe.CheckoutSession, error) {
	log.Infof("[Stripe engine] Processing money request with money: %v\n", money)
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
		ExpiresAt:  stripe.Int64(time.Now().Add(30 * time.Minute).Unix()),
	}

	log.Infoln("[Stripe engine] Creating checkout session with this reuest")
	checkoutSession, err := session.New(params)
	if err != nil {
		log.Errorf("[Stripe Engine]session.New: %v\n", err)
		return nil, err
	}
	log.Infoln("[Stripe engine] Create checkout session successful")
	return checkoutSession, nil
}

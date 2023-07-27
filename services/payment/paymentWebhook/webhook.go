package paymentwebhook

import (
	paymentservice "edaRestaurant/services/payment/paymentService"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v74/webhook"
)

type PamentWebHook struct {
	app            *fiber.App
	service        paymentservice.PaymentService
	endpointSecret string
}

func NewPaymentWebHook(service paymentservice.PaymentService) (PamentWebHook, error) {
	webhook := PamentWebHook{service: service}
	if err := webhook.initConnect(); err != nil {
		return webhook, err
	}
	return webhook, nil
}

func (s *PamentWebHook) initConnect() error {

	config := fiber.Config{
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}
	s.app = fiber.New(config)

	if err := s.initRoute(); err != nil {
		return err
	}
	endpointSecret := viper.GetViper().GetString("stripe_webhook_secret")
	if endpointSecret == "" {
		log.Errorf("Secret endpoint must specified\n")
		log.Exit(1)
	}
	s.endpointSecret = endpointSecret
	return nil
}

func (s *PamentWebHook) Run(port string) error {
	if err := s.app.Listen(port); err != nil {
		return err
	}
	return nil
}

func (s *PamentWebHook) initRoute() error {
	s.app.Post("/internal/payment-webhook", s.WebHookHandler())
	return nil
}

func (s *PamentWebHook) WebHookHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		event, err := webhook.ConstructEvent(ctx.Body(), ctx.GetReqHeaders()["Stripe-Signature"], s.endpointSecret)
		if err != nil {
			log.Printf("[Payment Webhook] Error: %v", err)
		}

		switch eventType := event.Type; eventType {
		case "checkout.session.completed":
			if err := s.service.HandleCompletedPayment(event.GetObjectValue("id")); err != nil {
				log.Errorf("[Payment Webhook] Error: %v\n", err)
				return err
			}
			log.Infof("[Payment Webhook] Completed\n")
		case "payment_intent.payment_failed":
			log.Infof("[Payment Webhook] Payment %v failed.\n", event.GetObjectValue("id"))
		case "checkout.session.expired":
			log.Infof("[Payment Webhook] Check out expired: %v\n", event.GetObjectValue("id"))
		default:
			log.Warnf("[Payment Webhook] Unhandled event: %v\n", eventType)
		}
		// TODO: handler incomming event, update repository and some relate stuff
		return nil
	}
}

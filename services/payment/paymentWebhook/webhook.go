package paymentwebhook

import (
	paymentservice "edaRestaurant/services/payment/paymentService"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v74/webhook"
)

type PamentWebHook struct {
	app     *fiber.App
	service paymentservice.PaymentService
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
		endpointSecret := "whsec_007a1cd0da122ad570a50dbeff82ebdeb8b9212cdde7891ae40827f17e21de9a"
		event, err := webhook.ConstructEvent(ctx.Body(), ctx.GetReqHeaders()["Stripe-Signature"], endpointSecret)
		if err != nil {
			log.Printf("[Webhook] Error: %v", err)
		}

		switch eventType := event.Type; eventType {
		case "checkout.session.completed":
			if err := s.service.HandleCompletedPayment(event.GetObjectValue("id")); err != nil {
				log.Printf("error: %v", err)
			}
			log.Println("Completed")
		case "payment_intent.payment_failed":
			log.Printf("payment %v failed.", event.GetObjectValue("id"))

		default:
			log.Printf("Unhandled event: %v", eventType)
		}
		// TODO: handler incomming event, update repository and some relate stuff
		return nil
	}
}

package main

import (
	"edaRestaurant/services/config"
	paymentservice "edaRestaurant/services/payment/paymentService"
	paymentwebhook "edaRestaurant/services/payment/paymentWebhook"
	"edaRestaurant/services/payment/paymentrepository"
	stripeengine "edaRestaurant/services/payment/stripeEngine"
	queueagent "edaRestaurant/services/queueAgent"
	"log"
)

func main() {
	repoconfig := config.NewMongoConfig(".")
	repo, err := paymentrepository.NewPaymentRepository(repoconfig)
	if err != nil {
		log.Fatal(err)
	}

	rabbitmqConfig := config.NewRabbitmqConfig(".")
	servicepublisher, err := queueagent.NewPublisher(rabbitmqConfig)
	stripeEngine := stripeengine.NewStripeEngine()

	service, err := paymentservice.NewPaymentService(repo, stripeEngine, servicepublisher, rabbitmqConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	go service.InitBackground()
	paymentWebhook, err := paymentwebhook.NewPaymentWebHook(service)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Println("running")
	if err := paymentWebhook.Run(":9123"); err != nil {
		log.Fatal(err)
	}
}

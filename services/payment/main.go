package main

import (
	"edaRestaurant/services/config"
	paymentservice "edaRestaurant/services/payment/paymentService"
	"edaRestaurant/services/payment/paymentrepository"
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

	service, err := paymentservice.NewPaymentService(repo, servicepublisher, rabbitmqConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	service.InitBackground()
	log.Println("running")
}

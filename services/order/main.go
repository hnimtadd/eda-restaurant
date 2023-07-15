package main

import (
	"edaRestaurant/services/config"
	orderrepo "edaRestaurant/services/order/orderRepo"
	orderservice "edaRestaurant/services/order/orderService"
	ordertransport "edaRestaurant/services/order/orderTransport"
	orderpublisher "edaRestaurant/services/queueAgent"
	"log"
)

func main() {
	repoConfig := config.NewMongoConfig(".")
	repo, err := orderrepo.NewOrderRepository(repoConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	rabbitmqConfig := config.NewRabbitmqConfig(".")
	servicepublisher, err := orderpublisher.NewPublisher(rabbitmqConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	service, err := orderservice.NewOrderService(servicepublisher, repo, rabbitmqConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	service.InitBackground()

	publisher, err := orderpublisher.NewPublisher(rabbitmqConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	transport, err := ordertransport.NewFiberTransport(service, publisher)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	transport.Run(":10077")
}

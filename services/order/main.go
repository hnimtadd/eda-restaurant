package main

import (
	"edaRestaurant/services/config"
	orderrepo "edaRestaurant/services/order/orderRepo"
	orderservice "edaRestaurant/services/order/orderService"
	ordertransport "edaRestaurant/services/order/orderTransport"
	orderpublisher "edaRestaurant/services/queueAgent"
	log "github.com/sirupsen/logrus"
)

func main() {
	repoConfig := config.NewMongoConfig(".")
	repo, err := orderrepo.NewOrderRepository(repoConfig)
	if err != nil {
		log.Errorf("error: %v", err)
	}

	rabbitmqConfig := config.NewRabbitmqConfig(".")
	servicepublisher, err := orderpublisher.NewPublisher(rabbitmqConfig)
	if err != nil {
		log.Errorf("error: %v", err)
	}
	log.Infoln("RabbitMq publihsher created")

	service, err := orderservice.NewOrderService(servicepublisher, repo, rabbitmqConfig)
	if err != nil {
		log.Errorf("error: %v", err)
	}
	log.Infoln("Order service initialized")
	go service.InitBackground()

	publisher, err := orderpublisher.NewPublisher(rabbitmqConfig)
	if err != nil {
		log.Errorf("error: %v", err)
	}
	transport, err := ordertransport.NewFiberTransport(service, publisher)
	if err != nil {
		log.Errorf("error: %v", err)
	}
	log.Infoln("Order fiber transport start")
	if err := transport.Run(":10077"); err != nil {
		log.Errorf("error: %v", err)
	}
}

package main

import (
	"edaRestaurant/services/config"
	queueagent "edaRestaurant/services/queueAgent"
	storagerepo "edaRestaurant/services/storage/storageRepo"
	storageservice "edaRestaurant/services/storage/storageService"
	"edaRestaurant/services/storage/transport"
	"log"
)

func main() {
	repoConfig := config.NewMongoConfig(".")
	rabbitmqConfig := config.NewRabbitmqConfig(".")
	repository, err := storagerepo.NewStorageRepository(repoConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	publisher, err := queueagent.NewPublisher(rabbitmqConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	service, err := storageservice.NewStorageService(repository, publisher, rabbitmqConfig)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	go service.InitBackground()
	transport, err := transport.NewFiberTransport(service, publisher)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if err := transport.Run(":9999"); err != nil {
		log.Fatalf("error: %v", err)
	}
}

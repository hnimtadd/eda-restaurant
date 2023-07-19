package main

import (
	"edaRestaurant/services/config"
	cookrepository "edaRestaurant/services/cook/cookRepository"
	cookservice "edaRestaurant/services/cook/cookService"
	queueagent "edaRestaurant/services/queueAgent"
	"log"
)

func main() {
	repoconfig := config.NewMongoConfig(".")
	repo, err := cookrepository.NewCookRepository(repoconfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	rabbitmqConfig := config.NewRabbitmqConfig(".")
	servicepublisher, err := queueagent.NewPublisher(rabbitmqConfig)

	service, err := cookservice.NewCookService(repo, servicepublisher, rabbitmqConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	service.InitBackground()
	log.Println("running")
}

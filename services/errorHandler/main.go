package main

import (
	"edaRestaurant/services/config"
	errorservice "edaRestaurant/services/errorHandler/errorService"
	queueagent "edaRestaurant/services/queueAgent"
	"log"
)

func main() {
	rabbitConfig := config.NewRabbitmqConfig(".")
	publisher, err := queueagent.NewPublisher(rabbitConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	service, err := errorservice.NewErrorService(publisher, rabbitConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	service.InitBackground()
}

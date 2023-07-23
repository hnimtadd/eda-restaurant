package main

import (
	"edaRestaurant/services/config"
	queueagent "edaRestaurant/services/queueAgent"
	tableservice "edaRestaurant/services/tableRunner/tableService"
	"log"
	"sync"
)

func main() {
	rabbitConfig := config.NewRabbitmqConfig(".")
	publisher, err := queueagent.NewPublisher(rabbitConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	service, err := tableservice.NewTableRunnerService(publisher, rabbitConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var (
		wg = sync.WaitGroup{}
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		service.InitBackground()
	}()
	log.Println("serving")
	wg.Wait()
}

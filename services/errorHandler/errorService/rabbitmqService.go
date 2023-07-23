package errorservice

import (
	"edaRestaurant/services/config"
	queueagent "edaRestaurant/services/queueAgent"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type errorService struct {
	conn      *amqp.Connection
	config    config.RabbitmqConfig
	publisher queueagent.Publisher
}

func NewErrorService(publisher queueagent.Publisher, config config.RabbitmqConfig) (ErrorService, error) {
	service := &errorService{
		config:    config,
		publisher: publisher,
	}
	if err := service.initConnection(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *errorService) initConnection() error {
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}
	s.conn = conn
	return nil
}

func (s *errorService) InitBackground() {
	s.ListenAndServeCookQueue()
}

func (s *errorService) ListenAndServeCookQueue() {
	ch, err := s.conn.Channel()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ch.ExchangeDeclare(
		"restaurant",
		"direct",
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueOverflowRejectPublishDLX,
		},
	)

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	queue, err := ch.QueueDeclare(
		"cook",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	ds, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for d := range ds {
			defer wg.Done()
			d.Ack(false)
		}
	}()

	log.Printf("[*] Listening to queue: %s\n", queue.Name)
	wg.Wait()

}

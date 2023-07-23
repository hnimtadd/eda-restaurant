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

	// err = ch.ExchangeDelete("errorEx", false, false)
	// if err != nil {
	// 	log.Fatalf("error: %v", err)
	// }
	err = ch.ExchangeDeclare(
		"errorEx",
		"fanout",
		true,
		false,
		true,
		false,
		amqp.Table{},
	)

	if err != nil {
		log.Println("Exchange config fail")
		log.Println("trying to redeclare exchange, unbind all other queue")
		log.Fatalf("error: %v", err)
	}

	queue, err := ch.QueueDeclare(
		"error",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ch.QueueBind(queue.Name, "", "errorEx", false, nil)
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
		defer wg.Done()
		for d := range ds {
			log.Printf("[ERROR]: received message on dead letter queue, from: %s, to: %s, corrId: %s, type:%s", d.ReplyTo, d.RoutingKey, d.CorrelationId, d.Type)
			err := s.HandlerErrorMessage(d)
			if err != nil {
				log.Printf("[ERROR]: %v", err)
				if d.Redelivered {
					d.Nack(false, false)
				} else {
					d.Nack(false, true)
				}
				continue
			}
			d.Ack(false)
		}
	}()

	log.Printf("[*] Listening to queue: %s\n", queue.Name)
	wg.Wait()
}

func (s *errorService) HandlerErrorMessage(d amqp.Delivery) error {
	msg := queueagent.PublishMessage{
		FromName: d.RoutingKey,
		ToName:   d.ReplyTo,
		CorrId:   d.CorrelationId,
		Type:     "reject",
		Body:     []byte("Message can't handler right now"),
	}
	err := s.publisher.PublishWithMessage(&msg)
	if err != nil {
		return err
	}
	return nil
}

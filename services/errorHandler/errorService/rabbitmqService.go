package errorservice

import (
	"edaRestaurant/services/config"
	queueagent "edaRestaurant/services/queueAgent"
	log "github.com/sirupsen/logrus"
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
		log.Errorf("[Error Service]: Error: %v", err)
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
	log.Infoln("[Error Service] Decalring channel")
	ch, err := s.conn.Channel()
	if err != nil {
		log.Errorf("error: %v", err)
	}

	log.Infoln("[Error Service] Declaring exclusive restaurant exchange")
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

	log.Infoln("[Error Service] Declaring dead-letter-exchange")
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
		log.Warnln("[Error Service] Exchange config fail")
		log.Warnln("[Error Service] Trying to redeclare exchange, unbind all other queue")
		log.Fatalf("[Error Service] Error: %v", err)
	}

	log.Infoln("[Error Service] Declaring dead-letter-queue associate with dead-letter-exchange")
	queue, err := ch.QueueDeclare(
		"error",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Errorf("[Error Service] error: %v", err)
	}

	log.Infoln("[Error Service] Bind dead-letter-queue to dead-letter-exchange")
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
			log.Infof("[Error Service]: received message on dead letter queue, from: %s, to: %s, corrId: %s, type:%s", d.ReplyTo, d.RoutingKey, d.CorrelationId, d.Type)
			err := s.HandlerErrorMessage(d)
			if err != nil {
				log.Errorf("[Error Service]: %v", err)
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

	log.Infof("[*] Listening to queue: %s\n", queue.Name)
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

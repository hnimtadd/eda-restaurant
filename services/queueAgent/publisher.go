package queueagent

import (
	"context"
	"edaRestaurant/services/config"
	"edaRestaurant/services/utils"
	"errors"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublishRequest struct {
	Type      string
	QueueName string
	Body      []byte
}
type Publisher interface {
	Publish(req *PublishRequest) error
}

type orderPublisher struct {
	conn   *amqp.Connection
	config config.RabbitmqConfig
}

func NewPublisher(config config.RabbitmqConfig) (Publisher, error) {
	publisher := &orderPublisher{
		config: config,
	}
	if err := publisher.initConnection(); err != nil {
		return publisher, err
	}
	return publisher, nil
}

func (s *orderPublisher) initConnection() error {
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}

	s.conn = conn
	return nil
}

func (s *orderPublisher) Publish(req *PublishRequest) error {
	ch, err := s.conn.Channel()

	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.Confirm(false); err != nil {
		return err
	}
	// q, err := ch.QueueDeclare( // reply queue
	// 	"",
	// 	false,
	// 	true,
	// 	true,
	// 	false,
	// 	nil,
	// )

	if err != nil {
		return err
	}

	corrId := utils.RandomString(32)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: continue implement publish method
	err = ch.PublishWithContext(
		ctx,
		"",
		req.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "plain/text",
			CorrelationId: corrId,
			Type:          req.Type,
			// ReplyTo:       q.Name,
			Body: req.Body,
		})

	var wg sync.WaitGroup

	result := make(chan error, 1)

	// listen to response

	chann := make(chan amqp.Confirmation, 1)
	chann = ch.NotifyPublish(chann)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			noti, ok := <-chann
			if !ok {
				log.Println("Debug point")
				continue
			}
			if noti.Ack {
				log.Println("ack")
				result <- nil
				return
			} else {
				log.Println("nack")
				result <- errors.New("Message can't publish right now")
				return
			}
		}
	}()
	for {
		res, ok := <-result
		log.Println("[RES]: ", res)
		if !ok {
			break
		}
		if res == nil {
			log.Printf("Published to queue: %s", req.QueueName)
		}
		return res
	}
	wg.Wait()
	return errors.New("Unreaced Code")
}

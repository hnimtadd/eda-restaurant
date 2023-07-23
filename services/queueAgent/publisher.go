package queueagent

import (
	"context"
	"edaRestaurant/services/config"
	"edaRestaurant/services/utils"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublishMessage struct {
	Type     string
	FromName string
	ToName   string
	Body     []byte
}

//	type PublishResponse struct {
//		Type          string
//		FromQueueName string
//		ToQueueName   string
//		Body          string
//	}
type Publisher interface {
	Publish(req *PublishMessage) error
	PublishWithValue(queueName string, reqType string, body interface{}) error
	PublishAndWaitForResponse(queueName string, reqType string, body interface{}) (PublishMessage, error)
	ReplyWithValue(d amqp.Delivery, reqType string, body interface{}) error
}

type publisher struct {
	conn   *amqp.Connection
	config config.RabbitmqConfig
}

func NewPublisher(config config.RabbitmqConfig) (Publisher, error) {
	publisher := &publisher{
		config: config,
	}
	if err := publisher.initConnection(); err != nil {
		return publisher, err
	}
	return publisher, nil
}

func (s *publisher) initConnection() error {
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}

	s.conn = conn
	return nil
}

func (s *publisher) Publish(req *PublishMessage) error {
	ch, err := s.conn.Channel()

	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.Confirm(false); err != nil {
		return err
	}
	if err != nil {
		return err
	}

	corrId := utils.RandomString(32)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: continue implement publish method
	log.Println(req)
	err = ch.PublishWithContext(
		ctx,
		"restaurant",
		req.ToName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "plain/text",
			CorrelationId: corrId,
			Type:          req.Type,
			Body:          req.Body,
			ReplyTo:       req.FromName,
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
			log.Printf("Published to queue: %s", req.ToName)
		}
		return res
	}
	wg.Wait()
	return errors.New("Unreaced Code")
}

func (s *publisher) PublishAndWaitForResponse(QueueName string, reqType string, value interface{}) (PublishMessage, error) {
	body, err := json.Marshal(&value)
	if err != nil {
		return PublishMessage{}, err
	}
	req := &PublishMessage{
		Type:   reqType,
		ToName: QueueName,
		Body:   body,
	}
	ch, err := s.conn.Channel()

	if err != nil {
		return PublishMessage{}, err
	}
	defer ch.Close()
	if err := ch.Confirm(false); err != nil {
		return PublishMessage{}, err
	}
	q, err := ch.QueueDeclare( // reply queue
		"",
		false,
		true,
		true,
		false,
		nil,
	)

	if err != nil {
		return PublishMessage{}, err
	}

	corrId := utils.RandomString(32)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: continue implement publish method
	err = ch.PublishWithContext(
		ctx,
		"restaurant",
		req.ToName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "plain/text",
			CorrelationId: corrId,
			Type:          req.Type,
			Body:          req.Body,
			ReplyTo:       q.Name,
		})

	var wg sync.WaitGroup

	result := make(chan amqp.Delivery, 1)

	// listen to response

	chann, err := ch.Consume(
		q.Name, //queuename
		"",     //consumer
		true,   //auto-ack
		true,   //exclusive
		false,  //no-local
		false,  //no-wait
		nil,    //args
	)
	// chann := make(chan amqp.Confirmation, 1)
	// chann = ch.NotifyPublish(chann)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range chann {
			log.Println("hello")
			if msg.CorrelationId == corrId {
				log.Println("Received reply message", msg)
				result <- msg
			}
		}
	}()
	for res := range result {
		log.Println("[RES]: ", res)
		rsp := PublishMessage{
			Type: res.Type,
			Body: res.Body,
		}
		return rsp, err
	}
	log.Println("Listening to response on: ", q.Name)
	wg.Wait()
	return PublishMessage{}, errors.New("Unreaced Code")
}

func (s *publisher) PublishWithValue(QueueName string, reqType string, value interface{}) error {
	body, err := json.Marshal(&value)
	if err != nil {
		return err
	}
	req := &PublishMessage{
		Type:   reqType,
		ToName: QueueName,
		Body:   body,
	}
	if err := s.Publish(req); err != nil {
		return err
	}
	return nil
}
func (s *publisher) ReplyWithValue(d amqp.Delivery, reqType string, body interface{}) error {
	ch, err := s.conn.Channel()

	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.Confirm(false); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bbody, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	// TODO: continue implement publish method
	err = ch.PublishWithContext(
		ctx,
		"restaurant",
		d.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:   "plain/text",
			CorrelationId: d.CorrelationId,
			Type:          reqType,
			Body:          bbody,
			// ReplyTo:       q.Name,
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
			log.Printf("Published to queue: %s", d.ReplyTo)
		}
		return res
	}
	wg.Wait()
	return errors.New("Unreaced Code")

}

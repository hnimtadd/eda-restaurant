package queueagent

import (
	"context"
	"edaRestaurant/services/config"
	"edaRestaurant/services/utils"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublishMessage struct {
	Type     string
	FromName string
	ToName   string
	Body     []byte
	CorrId   string
}
type Publisher interface {
	publish(req *PublishMessage) error
	PublishWithMessage(msg *PublishMessage) error
	PublishWithValue(fromName, toName, reqType string, body interface{}) error
	PublishAndWaitForResponse(fromName, toName, reqType string, body interface{}) (PublishMessage, error)
	MakeMessageWithValue(fromName, toName, reqType, corrId string, body interface{}) (PublishMessage, error)
	// ReplyWithValue(d amqp.Delivery, reqType string, body interface{}) error
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
	log.Infoln("[Queue Agent] connecting to rabbitMQ server")
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}
	s.conn = conn
	return nil
}

func (s *publisher) publish(req *PublishMessage) error {
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
	corrId := req.CorrId

	if corrId == "" {
		corrId = utils.RandomString(32)
	}
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
				log.Debugf("Some stuff need to debug")
				continue
			}
			if noti.Ack {
				log.Infof("Message acked")
				result <- nil
				return
			} else {
				log.Infof("Message nacked")
				result <- errors.New("Message can't publish right now")
				return
			}
		}
	}()
	for {
		res, ok := <-result
		if !ok {
			break
		}
		if res == nil {
			log.Infof("[Queue Agent] Published from %s, to %s, msg type: %s, corrId: %s\n", req.FromName, req.ToName, req.Type, req.CorrId)
		}
		return res
	}
	wg.Wait()
	return errors.New("Unreaced Code")
}

func (s *publisher) PublishAndWaitForResponse(fromName, toName string, reqType string, value interface{}) (PublishMessage, error) {
	body, err := json.Marshal(&value)
	if err != nil {
		return PublishMessage{}, err
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
		"reply",
		false,
		true,
		true,
		false,
		nil,
	)
	err = ch.QueueBind(q.Name, q.Name, "restaurant", false, nil)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

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
		toName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "plain/text",
			CorrelationId: corrId,
			Type:          reqType,
			Body:          body,
			ReplyTo:       q.Name,
		})

	if err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Printf("[Publisher] Published from %s, to %s, msg type: %s, corrId: %s\n", fromName, toName, reqType, corrId)
	var wg sync.WaitGroup

	result := make(chan amqp.Delivery, 1)

	// listen to response

	chann, err := ch.Consume(
		q.Name, //queuename
		"",     //consumer
		true,   //auto-ack
		false,  //exclusive
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
			if msg.CorrelationId == corrId {
				result <- msg
			}
		}
	}()
	for res := range result {
		rsp := PublishMessage{
			FromName: res.ReplyTo,
			ToName:   res.RoutingKey,
			Type:     res.Type,
			CorrId:   res.CorrelationId,
			Body:     res.Body,
		}
		log.Printf("[Publisher] Received reponse from: %s, to: %s, type: %s, corrrId: %s ", rsp.FromName, rsp.ToName, rsp.Type, rsp.CorrId)
		return rsp, err
	}
	log.Println("Listening to response on: ", q.Name)
	wg.Wait()
	return PublishMessage{}, errors.New("Unreaced Code")
}

func (s *publisher) PublishWithValue(fromName, toName, reqType string, value interface{}) error {
	body, err := json.Marshal(&value)
	if err != nil {
		return err
	}
	req := &PublishMessage{
		FromName: fromName,
		Type:     reqType,
		ToName:   toName,
		Body:     body,
	}
	if err := s.publish(req); err != nil {
		return err
	}
	return nil
}
func (s *publisher) PublishWithMessage(msg *PublishMessage) error {
	if err := s.publish(msg); err != nil {
		return err
	}
	return nil
}

func (s *publisher) MakeMessageWithValue(fromName, toName, reqType, corrId string, body interface{}) (PublishMessage, error) {
	bbody, err := json.Marshal(&body)
	if err != nil {
		return PublishMessage{}, err
	}
	msg := PublishMessage{
		FromName: fromName,
		ToName:   toName,
		Type:     reqType,
		CorrId:   corrId,
		Body:     bbody,
	}
	return msg, nil
}

// func (s *publisher) ReplyWithValue(d amqp.Delivery, reqType string, body interface{}) error {
// 	ch, err := s.conn.Channel()
//
// 	if err != nil {
// 		return err
// 	}
// 	defer ch.Close()
// 	if err := ch.Confirm(false); err != nil {
// 		return err
// 	}
//
// 	if err != nil {
// 		return err
// 	}
//
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	bbody, err := json.Marshal(&body)
// 	if err != nil {
// 		return err
// 	}
//
// 	// TODO: continue implement publish method
// 	err = ch.PublishWithContext(
// 		ctx,
// 		"restaurant",
// 		d.ReplyTo,
// 		false,
// 		false,
// 		amqp.Publishing{
// 			ContentType:   "plain/text",
// 			CorrelationId: d.CorrelationId,
// 			Type:          reqType,
// 			Body:          bbody,
// 			ReplyTo:       d.RoutingKey,
// 		})
//
// 	var wg sync.WaitGroup
//
// 	result := make(chan error, 1)
//
// 	// listen to response
//
// 	chann := make(chan amqp.Confirmation, 1)
// 	chann = ch.NotifyPublish(chann)
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		for {
// 			noti, ok := <-chann
// 			if !ok {
// 				log.Println("Debug point")
// 				continue
// 			}
// 			if noti.Ack {
// 				log.Println("ack")
// 				result <- nil
// 				return
// 			} else {
// 				log.Println("nack")
// 				result <- errors.New("Message can't publish right now")
// 				return
// 			}
// 		}
// 	}()
// 	for {
// 		res, ok := <-result
// 		log.Println("[RES]: ", res)
// 		if !ok {
// 			break
// 		}
// 		if res == nil {
// 			log.Printf("Published to queue: %s", d.RoutingKey)
// 		}
// 		return res
// 	}
// 	wg.Wait()
// 	return errors.New("Unreaced Code")
//
// }

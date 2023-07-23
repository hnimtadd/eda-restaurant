package tableservice

import (
	"edaRestaurant/services/config"
	queueagent "edaRestaurant/services/queueAgent"
	tableRunner "edaRestaurant/services/tableRunner/type"
	"encoding/json"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type tableRunnerService struct {
	conn      *amqp.Connection
	publisher queueagent.Publisher
	config    config.RabbitmqConfig
}

func NewTableRunnerService(publisher queueagent.Publisher, config config.RabbitmqConfig) (TableService, error) {
	service := &tableRunnerService{
		config:    config,
		publisher: publisher,
	}
	if err := service.initConnection(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *tableRunnerService) initConnection() error {
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}
	s.conn = conn
	return nil

}

func (s *tableRunnerService) InitBackground() {

	s.ListenAndServeCookQueue()
}
func (s *tableRunnerService) ListenAndServeCookQueue() {
	ch, err := s.conn.Channel()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	queue, err := ch.QueueDeclare(
		"tableRunner",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,
		queue.Name,
		"restaurant",
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
		defer wg.Done()
		for d := range ds {
			log.Println("[x] Received tablerunner msg: ")
			if d.Type == "serve" {
				if err := s.Serve(d.Body); err != nil {
					d.Nack(false, false)
				}
			} else if d.Type == "clean" {
				if err := s.Serve(d.Body); err != nil {
					d.Nack(false, false)
				}
			} else {
				log.Printf("[ERROR] No handler available")
				d.Nack(false, false)
			}
			d.Ack(false)
		}
	}()
	log.Printf("[*] Listening to queue: %s\n", queue.Name)
	wg.Wait()
}

func (s *tableRunnerService) Serve(msg []byte) error {
	log.Println("[tableRunnber] Serve food")
	var cReq tableRunner.TableServeRequest
	time.Sleep(time.Second * 5)
	if err := json.Unmarshal(msg, &cReq); err != nil {
		log.Printf("[Error]: %v", err)
		return err
	}
	log.Printf("[tableRunnber] Done serve food: %s \n", cReq.TableId)
	return nil
}

func (s *tableRunnerService) Clean(msg []byte) error {
	var cReq tableRunner.TableCleanRequest
	if err := json.Unmarshal(msg, &cReq); err != nil {
		log.Printf("[Error]: %v", err)
		return err
	}
	log.Printf("[tableRunnber] Clean food at : %s", cReq.TableId)
	time.Sleep(time.Second * 5)
	log.Printf("[tableRunnber] Done Clean food: %s \n", cReq.TableId)
	return nil
}

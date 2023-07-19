package orderservice

import (
	"edaRestaurant/services/config"
	"edaRestaurant/services/entities"
	orderrepo "edaRestaurant/services/order/orderRepo"
	"edaRestaurant/services/order/type"
	orderpublisher "edaRestaurant/services/queueAgent"
	queueagent "edaRestaurant/services/queueAgent"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
)

type orderService struct {
	conn      *amqp.Connection
	config    config.RabbitmqConfig
	repo      orderrepo.OrderRepository
	publisher queueagent.Publisher
}

func NewOrderService(publisher queueagent.Publisher, repo orderrepo.OrderRepository, config config.RabbitmqConfig) (OrderService, error) {
	service := &orderService{
		repo:      repo,
		config:    config,
		publisher: publisher,
	}
	if err := service.initConnection(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *orderService) initConnection() error {
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}
	s.conn = conn
	return nil
}

func (s *orderService) InitBackground() {
	s.ListenAndServeOrderQueue()
}

func (s *orderService) initQueue(queueName string) error {
	return nil
}

func (s *orderService) ListenAndServeOrderQueue() {
	ch, err := s.conn.Channel()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	queue, err := ch.QueueDeclare(
		"order",
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
			var order order.Order
			if err := json.Unmarshal(d.Body, &order); err != nil {
				if d.Redelivered {
					d.Nack(false, false)
					continue
				}
				d.Nack(false, true)
				continue
			}
			log.Printf("[x] create Order")
			if err := s.repo.CreateOrder(order); err != nil {
				d.Nack(false, true)
				continue
			}

			req := &orderpublisher.PublishRequest{
				Type:      "create",
				QueueName: "cook",
				Body:      d.Body,
			}
			if err := s.publisher.Publish(req); err != nil {
				log.Printf("error: %v", err)
			}
			d.Ack(false)
		}
	}()
	log.Println("[*] Listening to queue")
	wg.Wait()
}

func (s *orderService) GetOrderById(id string) (*entities.Order, error) {
	rorder, err := s.repo.GetOrderById(id)
	if err != nil {
		return nil, err
	}
	return rorder, nil
}

func (s *orderService) GetOrders() ([]entities.Order, error) {
	orders, err := s.repo.GetOrders()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

type checkIngredient struct {
	Id      string `json:"ingredient_id,omiempty"`
	Quality int    `json:"quality,omiempty"`
}

func (s *orderService) CreateDish(dish order.Dish) error {
	for _, ing := range dish.IngredientsId {
		agent := fiber.AcquireAgent()
		req := agent.Request()
		req.Header.SetMethod(fiber.MethodGet)
		log.Println("check ingredient: ", ing)
		bodycheck := checkIngredient{
			Id:      ing,
			Quality: 0,
		}
		byts, err := json.Marshal(&bodycheck)
		if err != nil {

		}
		req.SetBody(byts)
		req.SetRequestURI("http://localhost:9999/api/v1/event/check-ingredients")
		err = agent.Parse()
		if err != nil {
			return err
		}
		code, body, errs := agent.Bytes()
		if len(errs) > 0 {
			return errs[0]
		}
		log.Println(code, body)
		if code == fiber.StatusOK {
			time.Sleep(time.Second)
		} else if code == fiber.StatusNotFound {
			return errors.New("ingredient not found")
		}

	}
	if err := s.repo.CreateDish(dish); err != nil {
		return err
	}
	return nil
}

func (s *orderService) GetDishes() ([]entities.Dish, error) {
	dishes, err := s.repo.GetDishes()
	if err != nil {
		return nil, err
	}
	return dishes, nil
}

package cookservice

import (
	"edaRestaurant/services/config"
	cookrepository "edaRestaurant/services/cook/cookRepository"
	cook "edaRestaurant/services/cook/type"
	"edaRestaurant/services/entities"
	"edaRestaurant/services/order/type"
	queueagent "edaRestaurant/services/queueAgent"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
)

type cookService struct {
	conn      *amqp.Connection
	config    config.RabbitmqConfig
	repo      cookrepository.CookRepository
	publisher queueagent.Publisher
}

func NewCookService(repo cookrepository.CookRepository, publisher queueagent.Publisher, config config.RabbitmqConfig) (CookService, error) {
	service := &cookService{
		repo:      repo,
		config:    config,
		publisher: publisher,
	}
	if err := service.initConnection(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *cookService) initConnection() error {
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}
	s.conn = conn
	return nil
}

func (s *cookService) InitBackground() {
	s.ListenAndServeCookQueue()
}

func (s *cookService) ListenAndServeCookQueue() {
	ch, err := s.conn.Channel()
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
			if d.Type == "create" {
				var order order.Order
				if err := json.Unmarshal(d.Body, &order); err != nil {
					log.Printf("error: %v", err)
					d.Nack(true, false)
					continue
				}
				log.Printf("[x] Received cook order: tableId:  %v, dishId: [%v]", order.TableId, order.DishesId)
				if ok, msg, err := s.ServeOrder(order); !ok || err != nil {
					log.Printf("error: %v", err)
					if msg != nil {
						s.publisher.PublishWithMessage(msg)
						d.Ack(false)
					} else if !d.Redelivered {
						d.Nack(false, false)
					} else {
						d.Nack(false, true)
					}
					continue
				}
				d.Ack(false)
			} else {
				log.Printf("[cook] type not valid %s", d.Type)
				d.Nack(false, false)
			}
		}
	}()

	log.Printf("[*] Listening to queue: %s\n", queue.Name)
	wg.Wait()
}

func (s *cookService) ServeOrder(order order.Order) (bool, *queueagent.PublishMessage, error) {
	var (
		unservedDish = []string{}
		serveDish    = []string{}
	)
	for _, dishid := range order.DishesId {
		_, err := s.ServeDish(dishid)
		if err != nil {
			log.Printf("error: %v", err)
			unservedDish = append(unservedDish, dishid)
			continue
		}
		serveDish = append(serveDish, dishid)
	}

	if len(serveDish) > 0 {
		tableReq := cook.TableServeRequest{
			OrderId: order.OrderId,
			TableId: order.TableId,
			DishId:  serveDish,
		}
		// dish cooked, publish to tableRunner
		if err := s.publisher.PublishWithValue("order", "tableRunner", "serve", &tableReq); err != nil {
			log.Printf("error: %v", err)
			return false, nil, err
		}
		log.Printf("Served %d dish: [%v]", len(serveDish), serveDish)
	}

	if len(unservedDish) != 0 {
		order.DishesId = unservedDish
		body, err := json.Marshal(order)
		if err != nil {
			log.Printf("error: %v", err)
			return false, nil, err
		}
		req := queueagent.PublishMessage{
			Type:     "create",
			FromName: "cook",
			ToName:   "cook",
			Body:     body,
		}
		log.Printf("Republish unserved dish: %v", unservedDish)
		return false, &req, err
		// if err := s.publisher.Publish(req); err != nil {
		// 	log.Printf("error: %v", err)
		// }
	}
	log.Printf("order %v cooked\n", order)
	return true, nil, nil
}

func (s *cookService) ServeDish(dishid string) (entities.Dish, error) {
	var res entities.Dish
	dish, err := s.repo.GetDishWithId(dishid)
	if err != nil {
		return res, err
	}
	for _, ing := range dish.Ingredients {
		agent := fiber.AcquireAgent()
		req := agent.Request()
		req.Header.SetMethod(fiber.MethodGet)
		req.SetRequestURI("http://localhost:9999/api/v1/event/get-ingredients")
		err := agent.Parse()
		if err != nil {
			return res, err
		}
		code, body, errs := agent.Bytes()
		if len(errs) > 0 {
			return res, errs[0]
		}
		res := []entities.Ingredient{}
		// log.Println(code, body)
		if (200 <= code) && (code <= 300) {
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println("Get ingredient")
			}
			time.Sleep(time.Second)
			log.Printf("Take out ingredient %v \n, num: %v", ing, 1)
		} else {
		}
	}
	dish.CreatedAt = time.Now().Unix()
	return dish, nil
}

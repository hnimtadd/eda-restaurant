package storageservice

import (
	"edaRestaurant/services/config"
	"edaRestaurant/services/entities"
	queueagent "edaRestaurant/services/queueAgent"
	storagerepo "edaRestaurant/services/storage/storageRepo"
	storage "edaRestaurant/services/storage/type"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type storageService struct {
	conn      *amqp.Connection
	config    config.RabbitmqConfig
	repo      storagerepo.StorageRepo
	publisher queueagent.Publisher
}

func NewStorageService(repo storagerepo.StorageRepo, publisher queueagent.Publisher, config config.RabbitmqConfig) (StorageService, error) {
	service := &storageService{
		config:    config,
		repo:      repo,
		publisher: publisher,
	}
	if err := service.initConnection(); err != nil {
		return service, err
	}
	return service, nil
}

func (s *storageService) initConnection() error {
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}
	s.conn = conn
	return nil
}

func (s *storageService) InitBackground() {
	s.ListenAndServeQueue()
}

func (s *storageService) ListenAndServeQueue() {
	ch, err := s.conn.Channel()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	queue, err := ch.QueueDeclare(
		"storage",
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
			ok, msg, err := s.Handler(d)
			if err != nil {
				if !d.Redelivered {
					d.Nack(false, true)
					continue
				}
			}
			if msg != nil {
				s.publisher.Publish(msg)
			}
			if ok {
				d.Ack(false)
			}
		}
	}()
	log.Println("[*] Listening to queue")
	wg.Wait()
}

func (s *storageService) Handler(msg amqp.Delivery) (bool, *queueagent.PublishMessage, error) {
	return true, nil, nil
}

func (s *storageService) InsertIngredient(ingredient storage.Ingedient) error {
	if err := s.repo.RegisterNewIngredient(ingredient); err != nil {
		return err
	}
	return nil
}

func (s *storageService) GetIngredients() ([]entities.Ingredient, error) {
	ings, err := s.repo.GetIngredients()
	if err != nil {
		return nil, err
	}
	return ings, nil
}

func (s *storageService) GetDishes() ([]entities.Dish, error) {
	ings, err := s.repo.GetDishes()
	if err != nil {
		return nil, err
	}
	return ings, nil
}

func (s *storageService) InsertDish(dish storage.Dish) error {
	if err := s.repo.RegisterNewDish(dish); err != nil {
		return err
	}
	return nil
}

func (s *storageService) CheckIngredientAvailable(id string, num int) (bool, error) {
	contain, err := s.repo.CheckIngredientAvailable(id, num)
	if err != nil {
		return false, err
	}
	return contain, nil
}

func (s *storageService) UpdateQuality(id string, num int) error {
	if err := s.repo.UpdateQuality(id, num); err != nil {
		return err
	}
	return nil
}

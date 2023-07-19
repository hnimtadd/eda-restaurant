package orderrepo

import (
	"context"
	"edaRestaurant/services/config"
	"edaRestaurant/services/entities"
	"edaRestaurant/services/order/type"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type orderRepository struct {
	db     *mongo.Database
	config config.MongoConfig
}

func NewOrderRepository(config config.MongoConfig) (OrderRepository, error) {
	repo := &orderRepository{
		config: config,
	}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (repo *orderRepository) initDB() error {
	// TODO: init connection to mongodb
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOpts := options.Client().ApplyURI(repo.config.Source).SetAuth(
		options.Credential{
			AuthSource: repo.config.AuthSource,
			Username:   repo.config.Username,
			Password:   repo.config.Password,
		},
	).SetTLSConfig(nil).SetTimeout(5 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	repo.db = client.Database(repo.config.Database)
	log.Printf("Connected to mongodb with config: %v\n", repo.config)
	return nil
}

func (repo *orderRepository) CreateOrder(order order.Order) error {
	// TODO: add order to repository
	var orderEntity entities.Order
	orderEntity.OrderId = uuid.New().String()
	orderEntity.DishesId = order.DishesId

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := repo.db.Collection("orders").InsertOne(ctx, orderEntity)
	if err != nil {
		return err
	}
	return nil
}

func (repo *orderRepository) GetOrders() ([]entities.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cur, err := repo.db.Collection("orders").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	orders := []entities.Order{}
	for cur.Next(ctx) {
		var order entities.Order
		if err := cur.Decode(&order); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
func (repo *orderRepository) GetOrderById(id string) (*entities.Order, error) {
	// TODO: get order
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	filter := bson.D{primitive.E{Key: "orderid", Value: id}}
	res := repo.db.Collection("orders").FindOne(ctx, filter)
	if err := res.Err(); err != nil {
		return nil, err
	}
	var order entities.Order
	if err := res.Decode(&order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (repo *orderRepository) GetDishes() ([]entities.Dish, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	cur, err := repo.db.Collection("dishes").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	dishes := []entities.Dish{}
	for cur.Next(ctx) {
		var dish entities.Dish
		if err := cur.Decode(&dish); err != nil {
			return nil, err
		}
		dishes = append(dishes, dish)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return dishes, nil
}

func (repo *orderRepository) CreateDish(dish order.Dish) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	entityDish := entities.Dish{
		DishId:      uuid.New().String(),
		Description: dish.Description,
		Name:        dish.Name,
		Ingredients: dish.IngredientsId,
	}
	_, err := repo.db.Collection("dishes").InsertOne(ctx, entityDish)
	if err != nil {
		return err
	}
	return nil
}

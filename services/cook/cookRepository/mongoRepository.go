package cookrepository

import (
	"context"
	"edaRestaurant/services/config"
	cook "edaRestaurant/services/cook/type"
	"edaRestaurant/services/entities"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type cookRepository struct {
	db     *mongo.Database
	config config.MongoConfig
	// AddCookHistory(cook cook.Cook) error
	// GetCookHistories() ([]cook.Cook, error)
}

func NewCookRepository(config config.MongoConfig) (CookRepository, error) {
	repo := &cookRepository{
		config: config,
	}
	if err := repo.initDB(); err != nil {
		return repo, err
	}
	return repo, nil
}

func (repo *cookRepository) initDB() error {
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

func (repo *cookRepository) AddCookHistory(cook cook.Cook) error {
	// TODO: add cook history from cook db
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if _, err := repo.db.Collection("cooks").InsertOne(ctx, cook); err != nil {
		return err
	}
	return nil
}
func (repo *cookRepository) GetCookHistories() ([]cook.Cook, error) {
	// TODO: get cook histories from cook db
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cur, err := repo.db.Collection("cooks").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	cooks := []cook.Cook{}
	for cur.Next(ctx) {
		var cook cook.Cook
		if err := cur.Decode(&cook); err != nil {
			return nil, err
		}
		cooks = append(cooks, cook)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return cooks, nil
}

func (repo *cookRepository) GetDishWithId(id string) (entities.Dish, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	filter := bson.D{primitive.E{Key: "dishid", Value: id}}
	res := repo.db.Collection("dishes").FindOne(ctx, filter)
	if err := res.Err(); err != nil {
		return entities.Dish{}, err
	}
	var dish entities.Dish
	if err := res.Decode(&dish); err != nil {
		return entities.Dish{}, err
	}
	return dish, nil
}

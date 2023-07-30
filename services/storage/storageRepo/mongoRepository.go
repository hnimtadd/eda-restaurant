package storagerepo

import (
	"context"
	"edaRestaurant/services/config"
	"edaRestaurant/services/entities"
	storage "edaRestaurant/services/storage/type"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type storageRepository struct {
	db     *mongo.Database
	config config.MongoConfig
	// AddCookHistory(cook cook.Cook) error
	// GetCookHistories() ([]cook.Cook, error)
}

func NewStorageRepository(config config.MongoConfig) (StorageRepo, error) {
	repo := &storageRepository{
		config: config,
	}
	if err := repo.initDB(); err != nil {
		return repo, err
	}
	return repo, nil

}
func (repo *storageRepository) initDB() error {
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

func (repo *storageRepository) RegisterNewIngredient(ing storage.Ingedient) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	input := entities.Ingredient{
		Name:         ing.Name,
		IngredientId: uuid.New().String(),
		Quality:      ing.Quality,
	}
	defer cancel()
	if _, err := repo.db.Collection("ingredients").InsertOne(ctx, input); err != nil {
		return err
	}
	return nil
}

func (repo *storageRepository) GetDishes() ([]entities.Dish, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	cur, err := repo.db.Collection("dishes").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
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

func (repo *storageRepository) RegisterNewDish(dish storage.Dish) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	input := entities.Dish{
		Name:        dish.Name,
		DishId:      uuid.New().String(),
		Ingredients: dish.Ingredients,
		Description: dish.Description,
	}
	defer cancel()
	if _, err := repo.db.Collection("dishes").InsertOne(ctx, input); err != nil {
		return err
	}
	return nil
}

func (repo *storageRepository) CheckIngredientsAvailable(ingredients ...storage.Ingedient) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for _, ingredient := range ingredients {
		ok, err := repo.CheckIngredientAvailable(ctx, ingredient.Id, ingredient.Quality)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, errors.New(fmt.Sprintf("Ingredients (%v) is not available", ingredient.Id))
		}
	}
	return true, nil
}
func (repo *storageRepository) CheckIngredientAvailable(ctx context.Context, id string, num int) (bool, error) {
	log.Infof("[Storage Repository] Checking ingredient available for (%v)", id)
	filter := bson.D{primitive.E{Key: "ingredientid", Value: id}}
	cur := repo.db.Collection("ingredients").FindOne(ctx, filter)
	if err := cur.Err(); err != nil {
		return false, err
	}
	var ing entities.Ingredient
	if err := cur.Decode(&ing); err != nil {
		return false, err
	}
	if ing.Quality > num {
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("Ingredient %v is not enough with value: %v", id, num))
}

func (repo *storageRepository) UpdateQuality(ingredients ...storage.Ingedient) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for _, ingredient := range ingredients {
		if err := repo.UpdateOneIngredient(ctx, ingredient.Id, ingredient.Quality); err != nil {
			return err
		}
	}
	return nil
}
func (repo *storageRepository) UpdateOneIngredient(ctx context.Context, id string, num int) error {
	if id == "" {
		return errors.New("Id must not nil")
	}
	filter := bson.D{primitive.E{Key: "ingredientid", Value: id}}
	log.Infof("[Storage Repository] Update ingredient with id (%v) quality (%v)\n", id, num)
	update := bson.D{primitive.E{Key: "$inc", Value: bson.D{primitive.E{Key: "quality", Value: num}}}}
	cur := repo.db.Collection("ingredients").FindOneAndUpdate(ctx, filter, update)
	if err := cur.Err(); err != nil {
		log.Errorf("[Storage Repository] Error: %v\n", err)
		return err
	}
	return nil
}

func (repo *storageRepository) GetIngredients() ([]entities.Ingredient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cur, err := repo.db.Collection("ingredients").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	ingredients := []entities.Ingredient{}
	for cur.Next(ctx) {
		var ingredient entities.Ingredient
		if err := cur.Decode(&ingredient); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ingredient)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return ingredients, nil
}
func (repo *storageRepository) GetIngredientById(id string) (*entities.Ingredient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	filter := bson.D{primitive.E{Key: "ingredientid", Value: id}}
	cur := repo.db.Collection("ingredients").FindOne(ctx, filter)
	if err := cur.Err(); err != nil {
		return nil, err
	}
	var result entities.Ingredient
	if err := cur.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

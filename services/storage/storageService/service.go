package storageservice

import (
	"edaRestaurant/services/entities"
	storage "edaRestaurant/services/storage/type"
)

type StorageService interface {
	InitBackground()
	ListenAndServeQueue()
	InsertIngredient(ingredient storage.Ingedient) error
	GetIngredients() ([]entities.Ingredient, error)
	GetDishes() ([]entities.Dish, error)
	InsertDish(dish storage.Dish) error
	CheckIngredientAvailable(id string, num int) (bool, error)
	UpdateQuality(id string, num int) error
}

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
	GetIngredientById(id string) (*entities.Ingredient, error)
	GetDishes() ([]entities.Dish, error)
	InsertDish(dish storage.Dish) error
	CheckIngredientsAvailable(ingredients ...storage.Ingedient) (bool, error)
	UpdateQuality(ingredients ...storage.Ingedient) error
}

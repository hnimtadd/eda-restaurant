package storagerepo

import (
	"edaRestaurant/services/entities"
	storage "edaRestaurant/services/storage/type"
)

type StorageRepo interface {
	RegisterNewIngredient(ing storage.Ingedient) error

	CheckIngredientAvailable(id string, num int) (bool, error)
	UpdateQuality(id string, num int) error

	GetIngredients() ([]entities.Ingredient, error)
	RegisterNewDish(dish storage.Dish) error
	GetDishes() ([]entities.Dish, error)
}

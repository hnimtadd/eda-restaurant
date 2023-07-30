package storagerepo

import (
	"edaRestaurant/services/entities"
	storage "edaRestaurant/services/storage/type"
)

type StorageRepo interface {
	RegisterNewIngredient(ing storage.Ingedient) error

	CheckIngredientsAvailable(ingredients ...storage.Ingedient) (bool, error)
	UpdateQuality(ingredients ...storage.Ingedient) error

	GetIngredients() ([]entities.Ingredient, error)
	GetIngredientById(id string) (*entities.Ingredient, error)
	RegisterNewDish(dish storage.Dish) error
	GetDishes() ([]entities.Dish, error)
}

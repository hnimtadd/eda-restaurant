package cookrepository

import (
	cook "edaRestaurant/services/cook/type"
	"edaRestaurant/services/entities"
)

type CookRepository interface {
	AddCookHistory(cook cook.Cook) error
	GetCookHistories() ([]cook.Cook, error)
	GetDishWithId(id string) (entities.Dish, error)
}

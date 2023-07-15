package orderrepo

import (
	"edaRestaurant/services/entities"
	"edaRestaurant/services/order/type"
)

type OrderRepository interface {
	CreateOrder(order.Order) error
	GetOrders() ([]entities.Order, error)
	GetOrderById(string) (*entities.Order, error)
	GetDishes() ([]entities.Dish, error)
	CreateDish(order.Dish) error
}

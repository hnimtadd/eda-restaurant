package orderservice

import (
	"edaRestaurant/services/entities"
	order "edaRestaurant/services/order/type"
)

type OrderService interface {
	InitBackground()

	ListenAndServeOrderQueue()
	GetOrders() ([]entities.Order, error)
	GetOrderById(string) (*entities.Order, error)
	CreateDish(order.Dish) error
	GetDishes() ([]entities.Dish, error)
}

package orderservice

import (
	"edaRestaurant/services/entities"
)

type OrderService interface {
	InitBackground()

	ListenAndServeOrderQueue()
	GetOrders() ([]entities.Order, error)
	GetOrderById(string) (*entities.Order, error)
}

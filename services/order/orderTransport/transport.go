package ordertransport

import (
	orderservice "edaRestaurant/services/order/orderService"
	order "edaRestaurant/services/order/type"
	orderpublisher "edaRestaurant/services/queueAgent"
	queueagent "edaRestaurant/services/queueAgent"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/gofiber/fiber/v2"
)

type FiberTransport struct {
	app       *fiber.App
	service   orderservice.OrderService
	publisher queueagent.Publisher
}

func NewFiberTransport(service orderservice.OrderService, publisher queueagent.Publisher) (FiberTransport, error) {
	s := FiberTransport{
		service:   service,
		publisher: publisher,
	}
	if err := s.initConnection(); err != nil {
		return s, err
	}
	if err := s.initRoute(); err != nil {
		return s, err
	}
	return s, nil
}

func (s *FiberTransport) Run(port string) error {
	for _, route := range s.app.GetRoutes() {
		log.Infof("[%s] - %s\n", route.Method, route.Path)
	}
	if err := s.app.Listen(port); err != nil {
		return err
	}
	return nil
}

func (s *FiberTransport) initConnection() error {
	config := fiber.Config{
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}
	s.app = fiber.New(config)
	s.initRoute()
	return nil
}

func (s *FiberTransport) initRoute() error {
	s.app.Get("/api/v1/event/get-order/:id", s.GetOrderById())
	s.app.Get("/api/v1/event/get-orders", s.GetOrders())
	s.app.Post("/api/v1/event/create-order", s.CreateOrder())
	s.app.Get("/api/v1/event/get-dishes", s.GetDishes())
	s.app.Post("/api/v1/event/create-dish", s.CreateDish())
	s.app.Put("/api/v1/event/check-payment", s.CheckPayment())
	s.app.Put("/api/v1/event/make-payment", s.MakePayment())
	return nil
}

func (s *FiberTransport) GetOrderById() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id := ctx.Params("id")
		order, err := s.service.GetOrderById(id)
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(&order)
		return ctx.SendStatus(fiber.StatusOK)
	}
}

type OrderServiceResponse struct {
	Msg      string `json:"msg,omiempty"`
	Metadata []byte `json:"metadata"`
}

func (s *FiberTransport) CreateOrder() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var order order.Order
		rsp := OrderServiceResponse{}

		if err := json.Unmarshal(ctx.Body(), &order); err != nil {
			rsp.Msg = err.Error()
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(&rsp)
			return ctx.SendStatus(fiber.StatusBadRequest)
		}
		if order.TableId == "" {
			rsp.Msg = "tableId must not null"
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(&rsp)
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		body, err := json.Marshal(order)
		if err != nil {
			rsp.Msg = err.Error()
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(&rsp)
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		request := &orderpublisher.PublishMessage{
			ToName:   "order",
			FromName: "order",
			Body:     body,
		}
		if err := s.publisher.PublishWithMessage(request); err != nil {
			rsp.Msg = err.Error()
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(&rsp)
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		rsp.Msg = "Create Order success"
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(&rsp)
		return ctx.SendStatus(fiber.StatusOK)
	}
}

func (s *FiberTransport) GetOrders() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		orders, err := s.service.GetOrders()
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(orders)
		return ctx.SendStatus(fiber.StatusOK)

	}
}

func (s *FiberTransport) GetDishes() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		dishes, err := s.service.GetDishes()
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(&dishes)
		return ctx.SendStatus(fiber.StatusOK)
	}
}

func (s *FiberTransport) CreateDish() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var dish order.Dish
		if err := json.Unmarshal(ctx.Body(), &dish); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusBadRequest)
		}
		if err := s.service.CreateDish(dish); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": "success"})
		return ctx.SendStatus(fiber.StatusOK)

	}
}

func (s *FiberTransport) CleanTable() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id := ctx.Params("id")
		if err := s.service.CleanTable(id); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": "success"})
		return ctx.SendStatus(fiber.StatusOK)
	}
}

func (s *FiberTransport) CheckPayment() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req order.CheckPaymentRequest
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusBadRequest)
		}
		rsp, err := s.service.CheckPayment(req)
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusBadRequest)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(rsp)
		return ctx.SendStatus(fiber.StatusOK)
	}
}

func (s *FiberTransport) MakePayment() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req order.PaymentRequest
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusBadRequest)
		}
		rsp, err := s.service.MakePayment(req)
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(rsp)
		return ctx.SendStatus(fiber.StatusOK)
	}
}

package transport

import (
	queueagent "edaRestaurant/services/queueAgent"
	storageservice "edaRestaurant/services/storage/storageService"
	storage "edaRestaurant/services/storage/type"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type fiberTransport struct {
	app       *fiber.App
	service   storageservice.StorageService
	publisher queueagent.Publisher
}

func NewFiberTransport(
	service storageservice.StorageService,
	publisher queueagent.Publisher,
) (StorageTransport, error) {
	s := &fiberTransport{
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

func (s *fiberTransport) Run(port string) error {
	for _, route := range s.app.GetRoutes() {
		log.Printf("[%s] - %s\n", route.Method, route.Path)
	}
	if err := s.app.Listen(port); err != nil {
		return err
	}
	return nil
}

func (s *fiberTransport) initConnection() error {
	config := fiber.Config{
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}
	s.app = fiber.New(config)
	s.initRoute()
	return nil
}

func (s *fiberTransport) initRoute() error {
	s.app.Post("/api/v1/event/import-ingredient", s.InsertIngredient())
	s.app.Get("/api/v1/event/get-ingredients", s.GetIngredients())
	s.app.Post("/api/v1/event/create-dish", s.InsertDish())
	s.app.Get("/api/v1/event/get-dishes", s.GetDishes())
	s.app.Get("/api/v1/event/check-ingredients", s.CheckIngredient())
	s.app.Get("/api/v1/event/take-ingredients", s.TakeIngredients())
	// s.app.Post("/api/v1/event/create-order", s.CreateOrder())
	return nil
}
func (s *fiberTransport) GetIngredients() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		log.Println("GetIngredients")
		ingredients, err := s.service.GetIngredients()
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(ingredients)
		return ctx.SendStatus(fiber.StatusOK)
	}
}
func (s *fiberTransport) InsertIngredient() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var ingredient storage.Ingedient
		if err := json.Unmarshal(ctx.Body(), &ingredient); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		if err := s.service.InsertIngredient(ingredient); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": "Insert success"})
		return ctx.SendStatus(fiber.StatusOK)
	}
}
func (s *fiberTransport) InsertDish() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var dish storage.Dish
		if err := json.Unmarshal(ctx.Body(), &dish); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		if err := s.service.InsertDish(dish); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": "Insert success"})
		return ctx.SendStatus(fiber.StatusOK)
	}
}

func (s *fiberTransport) GetDishes() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		dishes, err := s.service.GetDishes()
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		json.NewEncoder(ctx.Response().BodyWriter()).Encode(dishes)
		return ctx.SendStatus(fiber.StatusOK)
	}
}

type checkIngredientRequest struct {
	Id      string `json:"ingredient_id,omiempty" db:"ingredient_id,omiempty"`
	Quality int    `json:"quality,omiempty" db:"quality,omiempty"`
}

func (s *fiberTransport) CheckIngredient() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		log.Println("checkpoint")
		var req checkIngredientRequest
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		log.Println(req)
		ok, err := s.service.CheckIngredientAvailable(req.Id, req.Quality)
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		log.Println(ok)
		if !ok {
			ctx.Response().Header.Set("Content-Type", "application/json")
			return ctx.SendStatus(fiber.StatusNotFound)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		return ctx.SendStatus(fiber.StatusOK)
	}
}
func (s *fiberTransport) TakeIngredients() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req checkIngredientRequest
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		log.Println(req)
		err := s.service.UpdateQuality(req.Id, req.Quality)
		if err != nil {
			ctx.Response().Header.Set("Content-Type", "application/json")
			json.NewEncoder(ctx.Response().BodyWriter()).Encode(map[string]string{"msg": err.Error()})
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		ctx.Response().Header.Set("Content-Type", "application/json")
		return ctx.SendStatus(fiber.StatusOK)
	}
}

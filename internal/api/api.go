package api

import (
	"gateway/internal/api/middleware"
	"gateway/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Routers - структура для хранения зависимостей роутов
type Routers struct {
	Service     service.Service
	AuthService service.AuthService
}

// NewRouters - конструктор для настройки API
func NewRouters(r *Routers, token string) *fiber.App {
	app := fiber.New()

	// Настройка CORS (разрешенные методы, заголовки, авторизация)
	app.Use(cors.New(cors.Config{
		AllowMethods:  "GET, POST, PUT, DELETE",
		AllowHeaders:  "Accept, Authorization, Content-Type, X-CSRF-Token, X-REQUEST-ID",
		ExposeHeaders: "Link",
		MaxAge:        300,
	}))
	app.Use(logger.New())
	// Группа маршрутов для авторизации
	authGroup := app.Group("/v1")
	authGroup.Post("/register", r.AuthService.RegisterHandler)
	authGroup.Post("/login", r.AuthService.LoginHandler)
	// Группа маршрутов с авторизацией
	protectedGroup := app.Group("/v1", middleware.Authorization(token))
	protectedGroup.Post("/create_task", r.Service.CreateTask)
	protectedGroup.Get("/task/:id", r.Service.GetTask)

	return app
}

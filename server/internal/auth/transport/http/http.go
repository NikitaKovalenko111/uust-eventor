package http

import (
	"example-service/internal/services"
	"example-service/internal/transport/http/controllers"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type HTTP struct {
	app         *fiber.App
	controllers *controllers.Controllers
	//authMiddleware          func(c *fiber.Ctx) error
}

func Init(services *services.Services, logger *slog.Logger, app *fiber.App /*authMiddleware func(c *fiber.Ctx) error*/) *HTTP {
	return &HTTP{
		app:         app,
		controllers: controllers.Init(services, logger),
	}
}

func (http *HTTP) Start() {
	http.controllers.ExampleController.RegisterRoutes("example", http.app)
}

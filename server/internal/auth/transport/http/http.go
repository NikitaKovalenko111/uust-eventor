package http

import (
	"eventor/internal/auth/services"
	"eventor/internal/auth/transport/http/controllers"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type HTTP struct {
	app            *fiber.App
	controllers    *controllers.Controllers
	authMiddleware fiber.Handler
}

func Init(services *services.Services, logger *slog.Logger, app *fiber.App, authMiddleware fiber.Handler) *HTTP {
	return &HTTP{
		app:            app,
		controllers:    controllers.Init(services, logger),
		authMiddleware: authMiddleware,
	}
}

func (http *HTTP) Start() {
	http.controllers.RegisterRoutes(http.app, http.authMiddleware)
}

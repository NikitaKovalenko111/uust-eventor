package http

import (
	"eventor/internal/event/services"
	"eventor/internal/event/transport/http/controllers"
	"log/slog"

	file_storage "eventor/internal/platform/storage/files"

	"github.com/gofiber/fiber/v2"
)

type HTTP struct {
	app            *fiber.App
	controllers    *controllers.Controllers
	authMiddleware *fiber.Handler
}

func Init(services *services.Services, logger *slog.Logger, app *fiber.App, authMiddleware *fiber.Handler, fileStorage *file_storage.FileStorage) *HTTP {
	return &HTTP{
		app:            app,
		controllers:    controllers.Init(services, fileStorage, logger),
		authMiddleware: authMiddleware,
	}
}

func (http *HTTP) Start() {
	http.controllers.EventController.RegisterRoutes(http.app, "/api/v1/events", http.authMiddleware)
}

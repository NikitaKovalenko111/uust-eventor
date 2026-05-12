package controllers

import (
	"example-service/internal/services"
	example_controller "example-service/internal/transport/http/controllers/example"
	"log/slog"
)

type Controllers struct {
	logger            *slog.Logger
	ExampleController *example_controller.ExampleController
	// Controllers
}

func Init(services *services.Services, logger *slog.Logger) *Controllers {
	return &Controllers{
		logger:            logger,
		ExampleController: example_controller.Init(services.ExampleService),
		// Inits of controllers
	}
}

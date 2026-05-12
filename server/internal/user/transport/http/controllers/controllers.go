package controllers

import (
	"eventor/internal/user/services"
	user_controller "eventor/internal/user/transport/http/controllers/user"
	"log/slog"
)

type Controllers struct {
	logger         *slog.Logger
	UserController *user_controller.UserController
	// Controllers
}

func Init(services *services.Services, logger *slog.Logger) *Controllers {
	return &Controllers{
		logger:         logger,
		UserController: user_controller.Init(services.UserService),
		// Inits of controllers
	}
}

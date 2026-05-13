package controllers

import (
	"eventor/internal/auth/services"
	auth_controller "eventor/internal/auth/transport/http/controllers/auth"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type Controllers struct {
	AuthController *auth_controller.AuthController
}

func Init(services *services.Services, logger *slog.Logger) *Controllers {
	return &Controllers{
		AuthController: auth_controller.NewAuthController(services.AuthService, logger),
	}
}

func (c *Controllers) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
	authRouter := app.Group("/api/v1/auth")
	basicRouter := authRouter

	c.AuthController.RegisterRoutes(basicRouter, authMiddleware)
}

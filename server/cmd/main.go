package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	auth_module "eventor/internal/auth/app"
	token_service "eventor/internal/auth/services/usecase/token"
	token_repo "eventor/internal/auth/storage/repositories/token"
	"eventor/internal/platform/config"
	sl "eventor/internal/platform/logger"
	"eventor/internal/platform/middleware"
	"eventor/internal/platform/storage"
	user_module "eventor/internal/user"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.MustLoad()

	logger := sl.InitLogger(cfg.Env)
	loggerMiddleware := middleware.NewLogger(logger)

	storage := storage.Init(&cfg.Storage)
	db := storage.Connect()
	defer db.Close()

	app := fiber.New(fiber.Config{
		StrictRouting: true,
		WriteTimeout:  cfg.HTTPServer.Timeout,
		IdleTimeout:   cfg.HTTPServer.IdleTimeout,
	})

	app.Use(loggerMiddleware)

	tokenService := token_service.Init(token_repo.Init(storage.Db), &cfg.JWT)
	authMiddleware := middleware.NewJWTMiddleware(tokenService, logger)

	userModule, userServices := user_module.New(cfg, app, logger, storage, tok)

	authModule, authServices := auth_module.New(cfg, app, logger, storage)
	authModule.Run()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	logger.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	go app.Listen(cfg.HTTPServer.Address)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	app.Shutdown()
}

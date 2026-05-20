package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	auth_module "eventor/internal/auth/app"
	token_service "eventor/internal/auth/services/usecase/token"
	token_repo "eventor/internal/auth/storage/repositories/token"
	event_module "eventor/internal/event"
	"eventor/internal/platform/config"
	sl "eventor/internal/platform/logger"
	"eventor/internal/platform/middleware"
	"eventor/internal/platform/storage"
	user_module "eventor/internal/user"

	_ "eventor/docs/swagger"

	swaggerUI "github.com/swaggo/fiber-swagger"

	"github.com/gofiber/fiber/v2"
)

// @title           Eventor API
// @version         1.0
// @description     Platform for event management: authentication, user profiles, events with comments and registration
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.email   support@eventor.local
// @contact.url     https://eventor.local/support

// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT

// @host           localhost:8080
// @BasePath       /api/v1
// @schemes        http https

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Enter JWT token in format: `Bearer <your_token_here>`. Token obtained from /auth/login or /auth/register.

// @tag.name   auth
// @tag.description  Authentication and authorization: login, registration, token refresh

// @tag.name   users
// @tag.description  User profile management: get/update own profile, avatar operations

// @tag.name   users-admin
// @tag.description  Administrative user operations: update/delete any user (moderator only)

// @tag.name   events
// @tag.description  Event management: CRUD, comments, registration, images

// @tag.name   health
// @tag.description  Service health checks
func main() {
	cfg := config.MustLoad()

	logger := sl.InitLogger(cfg.Env)
	loggerMiddleware := middleware.NewLogger(logger)

	storage := storage.Init(cfg)
	db := storage.Connect()
	defer db.Close()

	fileStorage, err := storage.NewFileStorage()
	if err != nil {
		logger.Error("failed to initialize file storage", slog.Any("error", err))
		os.Exit(1)
	}
	storage.FileStorage = fileStorage

	app := fiber.New(fiber.Config{
		StrictRouting: true,
		WriteTimeout:  cfg.HTTPServer.Timeout,
		IdleTimeout:   cfg.HTTPServer.IdleTimeout,
	})

	app.Use(loggerMiddleware)

	tokenService := token_service.Init(token_repo.Init(storage.Db), &cfg.JWT)
	authMiddleware := middleware.NewJWTMiddleware(tokenService, logger)

	userModule, userServices := user_module.New(cfg, app, logger, storage, authMiddleware)
	userModule.Run()

	authModule := auth_module.New(cfg, app, logger, storage, userServices.UserService, authMiddleware, tokenService)
	authModule.Run()

	eventModule := event_module.New(cfg, app, &authMiddleware, logger, storage.Db, userServices.UserService, storage.FileStorage)
	eventModule.Run()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Swagger UI: http://localhost:8080/swagger/index.html
	app.Get("/swagger/*", swaggerUI.WrapHandler)

	// Опционально: редирект /swagger → /swagger/index.html
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html")
	})

	logger.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	go app.Listen(cfg.HTTPServer.Address)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit

	app.Shutdown()
}

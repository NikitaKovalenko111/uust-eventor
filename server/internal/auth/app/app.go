package auth_module

import (
	//"context"
	//"crypto/tls"

	"eventor/internal/auth/contracts/user_provider"
	"eventor/internal/auth/services"
	token_service "eventor/internal/auth/services/usecase/token"
	"eventor/internal/auth/storage/repositories"
	"eventor/internal/auth/transport/http"
	"eventor/internal/platform/config"
	"eventor/internal/platform/storage"
	"log/slog"

	//"gopkg.in/gomail.v2"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	http   *http.HTTP
	app    *fiber.App
	config *config.Config
}

func New(cfg *config.Config, app *fiber.App, logger *slog.Logger, storage *storage.Storage, userService user_provider.UserProvider, authMiddleware fiber.Handler, tokenService *token_service.TokenService) *App {
	module := "auth"

	repos := repositories.Init(storage.Db)

	services := services.Init(repos, cfg, userService, tokenService)

	logger.Info("Successfully inited services!", slog.String("module", module))

	http := http.Init(services, logger, app, authMiddleware)

	return &App{
		http:   http,
		config: cfg,
		app:    app,
	}
}

func (app *App) Run() {
	app.http.Start()
}

func (app *App) Stop() {
	app.app.Shutdown()
}

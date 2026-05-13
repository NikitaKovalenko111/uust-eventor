package user_module

import (
	"eventor/internal/platform/config"
	"eventor/internal/platform/storage"
	"eventor/internal/user/services"
	"log/slog"

	"eventor/internal/user/storage/repositories"
	"eventor/internal/user/transport/http"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	http   *http.HTTP
	app    *fiber.App
	config *config.Config
}

func New(cfg *config.Config, app *fiber.App, logger *slog.Logger, storage *storage.Storage, authMiddleware fiber.Handler) (*App, *services.Services) {
	module := "user"

	repos := repositories.Init(storage.Db)

	logger.Info("Successfully inited repositories!", slog.String("module", module))

	services := services.Init(repos, cfg, storage.FileStorage)

	logger.Info("Successfully inited services!", slog.String("module", module))

	http := http.Init(services, logger, app, authMiddleware)

	return &App{
		http:   http,
		config: cfg,
		app:    app,
	}, services
}

func (app *App) Run() {
	app.http.Start()
}

func (app *App) Stop() {
	app.app.Shutdown()
}

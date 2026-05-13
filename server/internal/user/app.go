package user_module

import (
	//"context"
	//"crypto/tls"
	"eventor/internal/platform/config"
	"eventor/internal/user/services"
	"eventor/internal/user/storage"
	"log/slog"

	//redisStorage "eventor/internal/user/storage/redis"
	"eventor/internal/user/storage/repositories"
	"eventor/internal/user/transport/http"

	//"gopkg.in/gomail.v2"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	http   *http.HTTP
	app    *fiber.App
	config *config.Config
}

func New(cfg *config.Config, app *fiber.App, logger *slog.Logger, storage *storage.Storage, authMiddleware fiber.Handler) (*App, *services.Services) {
	repos := repositories.Init(storage.Db)

	logger.Info("Successfully inited repositories!")

	services := services.Init(repos, cfg)

	logger.Info("Successfully inited services!")

	http := http.Init(services, logger, app, authMiddleware)

	return &App{
		http:   http,
		config: cfg,
		app:    app,
	}, services
}

func (app *App) Run() {
	app.http.Start()

	go app.app.Listen(app.config.HTTPServer.Address)
}

func (app *App) Stop() {
	app.app.Shutdown()
}

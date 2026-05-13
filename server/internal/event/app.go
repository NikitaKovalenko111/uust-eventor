package event_module

import (
	//"context"
	//"crypto/tls"
	"database/sql"
	"eventor/internal/platform/config"
	"eventor/internal/user/services"
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

func New(cfg *config.Config, app *fiber.App, authMiddleware *fiber.Handler, logger *slog.Logger, db *sql.DB) *App {
	repos := repositories.Init(db)

	logger.Info("Successfully inited repositories!")

	services := services.Init(repos, cfg)

	logger.Info("Successfully inited services!")

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

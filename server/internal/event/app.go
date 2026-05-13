package event_module

import (
	//"context"
	//"crypto/tls"
	"database/sql"
	"eventor/internal/event/contracts/user_provider"
	"eventor/internal/event/services"
	"eventor/internal/event/storage/repositories"
	"eventor/internal/event/transport/http"
	"eventor/internal/platform/config"
	"log/slog"

	//"gopkg.in/gomail.v2"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	http   *http.HTTP
	app    *fiber.App
	config *config.Config
}

func New(cfg *config.Config, app *fiber.App, authMiddleware *fiber.Handler, logger *slog.Logger, db *sql.DB, userProvider user_provider.UserProvider) *App {
	module := "event"

	repos := repositories.Init(db)

	logger.Info("Successfully inited repositories!", slog.String("module", module))

	services := services.Init(repos, cfg, userProvider)

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

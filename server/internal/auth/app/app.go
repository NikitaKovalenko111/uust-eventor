package auth_module

import (
	//"context"
	//"crypto/tls"

	auth_middleware "eventor/internal/auth/middleware"
	"eventor/internal/auth/services"
	"eventor/internal/auth/storage/repositories"
	"eventor/internal/auth/transport/http"
	"eventor/internal/platform/config"
	sl "eventor/internal/platform/logger"
	"eventor/internal/platform/middleware"
	"eventor/internal/platform/storage"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2/middleware/cors"

	//"gopkg.in/gomail.v2"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	http   *http.HTTP
	app    *fiber.App
	config *config.Config
}

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func New(cfg *config.Config, app *fiber.App) *App {
	logger := sl.InitLogger(cfg.Env)

	logger.Info("Logger is enabled")
	logger.Debug("Debug is enabled")

	storage := storage.Init(&cfg.Storage)
	logger.Info("Successfully inited storage!")

	storage.Connect()

	logger.Info("Successfully connected to database!")

	//redisClient, err := redisStorage.NewClient(context.Background(), cfg)

	/*if err != nil {
		panic("Couldn't connect to redis!")
	}*/

	repos := repositories.Init(storage.Db)

	logger.Info("Successfully inited repositories!")

	/*d := gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}*/

	services := services.Init(repos, cfg)

	//authMiddleware := middleware.NewAuth(cfg, services.TokenService)

	logger.Info("Successfully inited services!")

	swaggerCfg := swagger.Config{
		BasePath: "/api",
		FilePath: "./docs/swagger.json",
		Path:     "docs",
		Title:    "Swagger API Docs",
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
	}))
	app.Use(swagger.New(swaggerCfg))
	app.Use(middleware.NewLogger(logger))

	authMiddleware := auth_middleware.NewJWTMiddleware(services.TokenService, logger)

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

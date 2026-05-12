package app

import (
	//"context"
	//"crypto/tls"
	"example-service/internal/config"
	"example-service/internal/logger/sl"
	"example-service/internal/middleware"
	"example-service/internal/services"
	"example-service/internal/storage"

	//redisStorage "example-service/internal/storage/redis"
	"example-service/internal/storage/repositories"
	"example-service/internal/transport/http"

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
func New(cfg *config.Config) *App {
	logger := sl.InitLogger(cfg.Env)

	logger.Info("Logger is enabled")
	logger.Debug("Debug is enabled")

	db := storage.Connect(cfg)

	logger.Info("Successfully connected to database!")

	storage := storage.Init(db)

	logger.Info("Successfully inited storage!")

	storage.Prepare()

	logger.Info("Successfully prepared db!")

	//redisClient, err := redisStorage.NewClient(context.Background(), cfg)

	/*if err != nil {
		panic("Couldn't connect to redis!")
	}*/

	logger.Info("Successfully connected to redis!")

	repos := repositories.Init(db)

	logger.Info("Successfully inited repositories!")

	/*d := gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}*/

	services := services.Init(repos /*d,*/, cfg /*redisClient*/)

	//authMiddleware := middleware.NewAuth(cfg, services.TokenService)

	logger.Info("Successfully inited services!")

	app := fiber.New(fiber.Config{
		StrictRouting: true,
		WriteTimeout:  cfg.HTTPServer.Timeout,
		IdleTimeout:   cfg.HTTPServer.IdleTimeout,
	})

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

	http := http.Init(services, app /*authMiddleware*/)

	return &App{
		http:   http,
		config: cfg,
		app:    app,
	}
}

func (app *App) Run() {
	app.http.Start()

	go app.app.Listen(app.config.HTTPServer.Address)
}

func (app *App) Stop() {
	app.app.Shutdown()
}

package example_controller

import (
	example_service "example-service/internal/services/usecase/example"
	example_dto "example-service/internal/transport/http/dto/example"
	_ "example-service/internal/types"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type ExampleController struct {
	logger         *slog.Logger
	ExampleService *example_service.ExampleService
	// Here are services associated with the controller
}

func Init(exampleService *example_service.ExampleService) *ExampleController {
	return &ExampleController{
		ExampleService: exampleService,
	}
}

func (controller *ExampleController) RegisterRoutes(route string, app *fiber.App /*authMiddleware func(c *fiber.Ctx) error*/) {
	router := app.Group("/" + route)

	router.Get("/health", controller.HealthCheck)
}

// HealthCheck godoc
//
//	@Summary		Check health
//	@Description	Checks health of the server.
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	example_dto.HealthCheckResponse
//	@Router			/health [get]
func (controller *ExampleController) HealthCheck(c *fiber.Ctx) error {
	var response example_dto.HealthCheckResponse

	response = example_dto.HealthCheckResponse{
		Status: fiber.StatusOK,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

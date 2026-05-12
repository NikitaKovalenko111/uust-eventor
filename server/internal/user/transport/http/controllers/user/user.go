package user_controller

import (
	user_service "eventor/internal/user/services/usecase/user"
	user_dto "eventor/internal/user/transport/http/dto/user"
	_ "eventor/internal/user/types"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type UserController struct {
	logger      *slog.Logger
	UserService *user_service.UserService
	// Here are services associated with the controller
}

func Init(userService *user_service.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

func (controller *UserController) RegisterRoutes(route string, app *fiber.App /*authMiddleware func(c *fiber.Ctx) error*/) {
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
//	@Success		200	{object}	user_dto.HealthCheckResponse
//	@Router			/health [get]
func (controller *UserController) HealthCheck(c *fiber.Ctx) error {
	var response user_dto.HealthCheckResponse

	response = user_dto.HealthCheckResponse{
		Status: fiber.StatusOK,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// internal/user/transport/http/controller/user.go
package user_controller

import (
	"errors"
	"log/slog"
	"strconv"

	"eventor/internal/platform/types"
	domain_errors "eventor/internal/user/domain/errors"
	user_service "eventor/internal/user/services/usecase/user"
	user_dto "eventor/internal/user/transport/http/dto/user"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// UserController управляет только профилями пользователей
type UserController struct {
	logger      *slog.Logger
	validator   *validator.Validate
	userService *user_service.UserService
}

// New создаёт контроллер
func Init(logger *slog.Logger, userService *user_service.UserService) *UserController {
	return &UserController{
		logger:      logger,
		validator:   validator.New(),
		userService: userService,
	}
}

// RegisterRoutes регистрирует маршруты управления профилем
// Все маршруты защищены authMiddleware
func (c *UserController) RegisterRoutes(app *fiber.App, rout string, authMiddleware fiber.Handler) {
	router := app.Group(rout, authMiddleware)

	router.Get("/me", c.GetMe)      // Текущий пользователь
	router.Put("/me", c.UpdateMe)   // Обновление своего профиля
	router.Get("/:id", c.GetByID)   // Просмотр другого пользователя (опционально)
	router.Put("/:id", c.Update)    // Обновление другого (только для модераторов)
	router.Delete("/:id", c.Delete) // Удаление (только для модераторов/админов)

	hRouter := app.Group("/api/v1/users")

	hRouter.Get("/health", c.HealthCheck)
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

// =====================================================
// HANDLERS: Мой профиль
// =====================================================

// GetMe godoc
//
//	@Summary		Get current user profile
//	@Description	Returns profile of authenticated user
//	@Tags			users
//	@Produce		json
//	@Success		200	{object}	user_dto.UserResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/users/me [get]
func (c *UserController) GetMe(ctx *fiber.Ctx) error {
	// ID берём из контекста (устанавливается authMiddleware)
	userID, ok := ctx.Locals("user_id").(types.IdType)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error: "unauthorized",
			Code:  fiber.StatusUnauthorized,
		})
	}

	user, err := c.userService.GetByID(ctx.UserContext(), userID)
	if err != nil {
		if errors.Is(err, domain_errors.ErrUserNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: "user not found",
				Code:  fiber.StatusNotFound,
			})
		}
		return c.handleServiceError(ctx, err, "get me")
	}

	return ctx.Status(fiber.StatusOK).JSON(user_dto.ToResponse(user))
}

// UpdateMe godoc
//
//	@Summary		Update my profile
//	@Description	Update authenticated user's profile
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body	user_dto.UpdateUserRequest	true	"Profile update"
//	@Success		200		{object}	user_dto.UserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse	"Email exists"
//	@Security		BearerAuth
//	@Router			/api/v1/users/me [put]
func (c *UserController) UpdateMe(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(types.IdType)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Error: "unauthorized",
			Code:  fiber.StatusUnauthorized,
		})
	}

	var req user_dto.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid request body",
			Code:  fiber.StatusBadRequest,
		})
	}

	// Валидация
	if err := c.validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Code:    fiber.StatusBadRequest,
			Details: formatValidationErrors(err),
		})
	}

	// Запрещаем менять email/role через этот эндпоинт (только через админку или отдельный флоу)
	if req.Email != nil || req.Role != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Error: "email and role cannot be changed via this endpoint",
			Code:  fiber.StatusForbidden,
		})
	}

	updated, err := c.userService.Update(ctx.UserContext(), userID, &req)
	if err != nil {
		return c.handleServiceError(ctx, err, "update me")
	}

	return ctx.Status(fiber.StatusOK).JSON(user_dto.ToResponse(updated))
}

// =====================================================
// HANDLERS: Управление другими пользователями (для модераторов)
// =====================================================

// GetByID godoc
//
//	@Summary		Get user by ID
//	@Description	Returns user details (public fields only)
//	@Tags			users
//	@Produce		json
//	@Param			id	path	uint	true	"User ID"
//	@Success		200	{object}	user_dto.UserResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/users/{id} [get]
func (c *UserController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	idType := types.IdType(id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid user ID",
			Code:  fiber.StatusBadRequest,
		})
	}

	// Опционально: проверка прав (может ли текущий пользователь видеть этого)
	// currentRole := ctx.Locals("user_role").(string)
	// if currentRole != "moderator" && id != currentUserID { ... }

	user, err := c.userService.GetByID(ctx.UserContext(), idType)
	if err != nil {
		if errors.Is(err, domain_errors.ErrUserNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: "user not found",
				Code:  fiber.StatusNotFound,
			})
		}
		return c.handleServiceError(ctx, err, "get user")
	}

	return ctx.Status(fiber.StatusOK).JSON(user_dto.ToResponse(user))
}

// Update godoc
//
//	@Summary		Update user (admin)
//	@Description	Update any user's profile (moderator+ only)
//	@Tags			users-admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path	uint						true	"User ID"
//	@Param			request	body	user_dto.UpdateUserRequest	true	"Update data"
//	@Success		200		{object}	user_dto.UserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse	"Forbidden"
//	@Failure		404		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/users/{id} [put]
func (c *UserController) Update(ctx *fiber.Ctx) error {
	// Проверка прав: только модераторы
	role, ok := ctx.Locals("user_role").(string)
	if !ok || role != "moderator" {
		return ctx.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Error: "moderators only",
			Code:  fiber.StatusForbidden,
		})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	idType := types.IdType(id)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid user ID",
			Code:  fiber.StatusBadRequest,
		})
	}

	var req user_dto.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid request body",
			Code:  fiber.StatusBadRequest,
		})
	}

	if err := c.validator.Struct(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Code:    fiber.StatusBadRequest,
			Details: formatValidationErrors(err),
		})
	}

	updated, err := c.userService.Update(ctx.UserContext(), idType, &req)
	if err != nil {
		return c.handleServiceError(ctx, err, "admin update user")
	}

	return ctx.Status(fiber.StatusOK).JSON(user_dto.ToResponse(updated))
}

// Delete godoc
//
//	@Summary		Delete user (admin)
//	@Description	Permanently delete user (moderator+ only)
//	@Tags			users-admin
//	@Produce		json
//	@Param			id	path	uint	true	"User ID"
//	@Success		204	"Deleted"
//	@Failure		400	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/users/{id} [delete]
func (c *UserController) Delete(ctx *fiber.Ctx) error {
	role, ok := ctx.Locals("user_role").(string)
	if !ok || role != "moderator" {
		return ctx.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Error: "moderators only",
			Code:  fiber.StatusForbidden,
		})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	idType := types.IdType(id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid user ID",
			Code:  fiber.StatusBadRequest,
		})
	}

	if err := c.userService.Delete(ctx.UserContext(), idType); err != nil {
		if errors.Is(err, domain_errors.ErrUserNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: "user not found",
				Code:  fiber.StatusNotFound,
			})
		}
		return c.handleServiceError(ctx, err, "delete user")
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// =====================================================
// ERROR HANDLING & HELPERS
// =====================================================

func (c *UserController) handleServiceError(ctx *fiber.Ctx, err error, operation string) error {
	c.logger.Error("service error", "operation", operation, "error", err)

	switch {
	case errors.Is(err, domain_errors.ErrValidation):
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation failed",
			Code:    fiber.StatusBadRequest,
			Details: extractValidationErrorDetails(err),
		})
	case errors.Is(err, domain_errors.ErrEmailExists):
		return ctx.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Error: err.Error(),
			Code:  fiber.StatusConflict,
		})
	case errors.Is(err, domain_errors.ErrUserNotFound):
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "user not found",
			Code:  fiber.StatusNotFound,
		})
	default:
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "internal server error",
			Code:  fiber.StatusInternalServerError,
		})
	}
}

// ErrorResponse стандартный формат ошибки
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    int               `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// formatValidationErrors форматирует ошибки validator
func formatValidationErrors(err error) map[string]string {
	details := make(map[string]string)
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := e.Field()
			switch e.Tag() {
			case "required":
				details[field] = "is required"
			case "email":
				details[field] = "must be valid email"
			case "min":
				details[field] = "must be at least " + e.Param() + " characters"
			case "max":
				details[field] = "must not exceed " + e.Param() + " characters"
			default:
				details[field] = "invalid value"
			}
		}
	}
	return details
}

func extractValidationErrorDetails(err error) map[string]string {
	for err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(errs)
		}
		err = errors.Unwrap(err)
	}
	return nil
}

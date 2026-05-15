// internal/user/transport/http/controller/user.go
package user_controller

import (
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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
	publicRouter := app.Group(rout)

	router.Get("/me", c.GetMe)    // Текущий пользователь
	router.Put("/me", c.UpdateMe) // Обновление своего профиля
	router.Put("/me/avatar", c.UpdateAvatar)
	router.Delete("/me/avatar", c.DeleteAvatar)
	router.Get("/:id/avatar", c.GetAvatar)
	publicRouter.Get("/:id/avatar/file", c.GetAvatarFile)
	router.Get("/:id", c.GetByID)   // Просмотр другого пользователя (опционально)
	router.Put("/:id", c.Update)    // Обновление другого (только для модераторов)
	router.Delete("/:id", c.Delete) // Удаление (только для модераторов/админов)
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns service health status. Public endpoint.
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} user_dto.HealthCheckResponse "Service is healthy"
// @Router /health [get]
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
// @Summary Get my profile
// @Description Returns profile of the authenticated user
// @Tags users
// @Produce json
// @Success 200 {object} user_dto.UserResponse "Profile retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized: invalid or missing token"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/users/me [get]
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
// @Summary Update my profile
// @Description Update authenticated user's profile. Email and role cannot be changed via this endpoint.
// @Tags users
// @Accept json
// @Produce json
// @Param request body user_dto.UpdateUserRequest true "Profile fields to update (all optional)"
// @Success 200 {object} user_dto.UserResponse "Profile updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body or validation failed"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden: email/role change not allowed via this endpoint"
// @Failure 409 {object} ErrorResponse "Conflict: email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/users/me [put]
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

// UpdateAvatar godoc
// @Summary Update my avatar
// @Description Upload new avatar image for authenticated user. Accepts multipart/form-data.
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image file (JPEG, PNG, WebP)"
// @Success 200 {object} user_dto.UserResponse "Avatar updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid file or not an image"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "File storage error"
// @Security BearerAuth
// @Router /api/v1/users/me/avatar [put]
func (c *UserController) UpdateAvatar(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(types.IdType)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(ErrorResponse{Error: "unauthorized", Code: http.StatusUnauthorized})
	}

	fileHeader, err := ctx.FormFile("avatar")
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "avatar file is required", Code: http.StatusBadRequest})
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "avatar must be an image", Code: http.StatusBadRequest})
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.logger.Error("failed to open avatar file", slog.Any("error", err))
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "cannot read avatar file", Code: http.StatusBadRequest})
	}
	defer file.Close()

	updated, err := c.userService.SetAvatar(ctx.UserContext(), userID, file, fileHeader.Size, contentType)
	if err != nil {
		return c.handleServiceError(ctx, err, "update avatar")
	}

	return ctx.Status(http.StatusOK).JSON(user_dto.ToResponse(updated))
}

// DeleteAvatar godoc
// @Summary Delete my avatar
// @Description Remove avatar from authenticated user's profile
// @Tags users
// @Produce json
// @Success 200 {object} user_dto.UserResponse "Avatar deleted successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Avatar not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/users/me/avatar [delete]
func (c *UserController) DeleteAvatar(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(types.IdType)
	if !ok {
		return ctx.Status(http.StatusUnauthorized).JSON(ErrorResponse{Error: "unauthorized", Code: http.StatusUnauthorized})
	}

	updated, err := c.userService.DeleteAvatar(ctx.UserContext(), userID)
	if err != nil {
		return c.handleServiceError(ctx, err, "delete avatar")
	}

	return ctx.Status(http.StatusOK).JSON(user_dto.ToResponse(updated))
}

// GetAvatar godoc
// @Summary Get avatar as data URL
// @Description Returns avatar of specified user as base64 data URL. Requires authentication.
// @Tags users
// @Produce json
// @Param id path uint true "User ID" minimum(1)
// @Success 200 {object} AvatarResponse "Avatar data URL"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User or avatar not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/users/{id}/avatar [get]
func (c *UserController) GetAvatar(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "invalid user ID", Code: http.StatusBadRequest})
	}

	avatarURL, err := c.userService.GetAvatarURL(ctx.UserContext(), types.IdType(id))
	if err != nil {
		return c.handleServiceError(ctx, err, "get avatar")
	}
	c.logger.Info("avatar metadata requested", slog.Uint64("user_id", id), slog.String("avatar_url", avatarURL))

	reader, contentType, err := c.userService.OpenAvatar(ctx.UserContext(), types.IdType(id))
	if err != nil {
		return c.handleServiceError(ctx, err, "get avatar file")
	}
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	avatarBytes, err := io.ReadAll(reader)
	if err != nil {
		c.logger.Error("failed to read avatar file", slog.Any("error", err))
		return ctx.Status(http.StatusInternalServerError).JSON(ErrorResponse{Error: "cannot read avatar file", Code: http.StatusInternalServerError})
	}
	c.logger.Info("avatar file read", slog.Uint64("user_id", id), slog.String("content_type", contentType), slog.Int("bytes", len(avatarBytes)))

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	dataURL := "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(avatarBytes)
	c.logger.Info("avatar data url built", slog.Uint64("user_id", id), slog.Int("data_url_length", len(dataURL)))
	return ctx.Status(http.StatusOK).JSON(AvatarResponse{DataURL: dataURL, URL: avatarURL})
}

// GetAvatarFile godoc
// @Summary Get avatar as raw file
// @Description Download avatar image file directly. Public endpoint (no auth required).
// @Tags users
// @Produce octet-stream
// @Param id path uint true "User ID" minimum(1)
// @Success 200 {file} binary "Avatar image file"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 404 {object} ErrorResponse "User or avatar not found"
// @Failure 500 {object} ErrorResponse "File read error"
// @Router /api/v1/users/{id}/avatar/file [get]
func (c *UserController) GetAvatarFile(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(ErrorResponse{Error: "invalid user ID", Code: http.StatusBadRequest})
	}

	reader, contentType, err := c.userService.OpenAvatar(ctx.UserContext(), types.IdType(id))
	if err != nil {
		return c.handleServiceError(ctx, err, "get avatar file")
	}
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			_ = closer.Close()
		}
	}()
	c.logger.Info("avatar file stream requested", slog.Uint64("user_id", id), slog.String("content_type", contentType))

	ctx.Type(contentType)
	return ctx.SendStream(reader)
}

// =====================================================
// HANDLERS: Управление другими пользователями (для модераторов)
// =====================================================

// GetByID godoc
// @Summary Get user by ID
// @Description Returns public profile information of specified user
// @Tags users
// @Produce json
// @Param id path uint true "User ID" minimum(1)
// @Success 200 {object} user_dto.UserResponse "User profile retrieved"
// @Failure 400 {object} ErrorResponse "Invalid user ID format"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
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
// @Summary Update user profile (admin)
// @Description Update any user's profile. Requires moderator role. All fields optional.
// @Tags users-admin
// @Accept json
// @Produce json
// @Param id path uint true "User ID" minimum(1)
// @Param request body user_dto.UpdateUserRequest true "Fields to update (all optional)"
// @Success 200 {object} user_dto.UserResponse "Profile updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or validation failed"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden: moderators only"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 409 {object} ErrorResponse "Conflict: email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/users/{id} [put]
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
// @Summary Delete user (admin)
// @Description Permanently delete user account. Requires moderator role.
// @Tags users-admin
// @Produce json
// @Param id path uint true "User ID" minimum(1)
// @Success 204 "User deleted successfully (no content)"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden: moderators only"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/users/{id} [delete]
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
	case errors.Is(err, domain_errors.ErrAvatarNotFound):
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "avatar not found",
			Code:  fiber.StatusNotFound,
		})
	case errors.Is(err, domain_errors.ErrFileStorage):
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "avatar storage error",
			Code:  fiber.StatusInternalServerError,
		})
	default:
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "internal server error",
			Code:  fiber.StatusInternalServerError,
		})
	}
}

// ErrorResponse
// @Description Standard error response format
type ErrorResponse struct {
	// Human-readable error message
	// @example "user not found"
	Error string `json:"error"`
	// HTTP status code
	// @example 404
	Code int `json:"code"`
	// Detailed validation errors (if applicable)
	// @example {"email": "must be valid email", "name": "is required"}
	Details map[string]string `json:"details,omitempty"`
}

// AvatarResponse
// @Description Response containing avatar data (for GetAvatar endpoint)
type AvatarResponse struct {
	// Avatar as data URL (base64-encoded, for inline display)
	// @example "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
	DataURL string `json:"data_url,omitempty"`
	// Direct URL to fetch avatar file
	// @example "/api/v1/users/12345/avatar/file"
	URL string `json:"url,omitempty"`
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

package auth_controller

import (
	"errors"
	domain_errors "eventor/internal/auth/domain/errors"
	auth_service "eventor/internal/auth/services/usecase/auth"
	auth_dto "eventor/internal/auth/transport/http/dto/auth"
	"eventor/internal/platform/types"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	authService *auth_service.AuthService
	logger      *slog.Logger
}

func NewAuthController(authService *auth_service.AuthService, logger *slog.Logger) *AuthController {
	return &AuthController{
		authService: authService,
		logger:      logger,
	}
}

func (ac *AuthController) RegisterRoutes(basicRouter fiber.Router, protectedRouter fiber.Router) {
	basicRouter.Post("/register", ac.register)
	basicRouter.Post("/login", ac.login)
	protectedRouter.Post("/refresh", ac.refreshToken)
}

func (ac *AuthController) login(c *fiber.Ctx) error {
	var req auth_dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		ac.logger.Warn("failed to parse login request", slog.Any("error", err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	tokenPair, err := ac.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain_errors.ErrInvalidCredentials) {
			ac.logger.Info("login failed: invalid credentials", slog.String("email", req.Email))
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		ac.logger.Error("login error", slog.Any("error", err), slog.String("email", req.Email))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	ac.logger.Info("user logged in", slog.String("email", req.Email))
	return c.Status(http.StatusOK).JSON(auth_dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.String(),
	})
}

func (ac *AuthController) register(c *fiber.Ctx) error {
	var req auth_dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		ac.logger.Warn("failed to parse register request", slog.Any("error", err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	tokenPair, err := ac.authService.Register(req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, domain_errors.ErrEmailAlreadyExists) {
			ac.logger.Info("registration failed: email already exists", slog.String("email", req.Email))
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "email already registered"})
		}
		ac.logger.Error("registration error", slog.Any("error", err), slog.String("email", req.Email))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	ac.logger.Info("user registered", slog.String("email", req.Email), slog.String("role", req.Role))
	return c.Status(http.StatusCreated).JSON(auth_dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.String(),
	})
}

func (ac *AuthController) refreshToken(c *fiber.Ctx) error {
	var req auth_dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		ac.logger.Warn("failed to parse refresh token request", slog.Any("error", err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		ac.logger.Warn("user_id not found in context")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	userID, ok := userIDRaw.(types.IdType)
	if !ok {
		ac.logger.Warn("user_id has invalid type")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	emailRaw := c.Locals("email")
	if emailRaw == nil {
		ac.logger.Warn("email not found in context")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	email, ok := emailRaw.(string)
	if !ok {
		ac.logger.Warn("email has invalid type")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	roleRaw := c.Locals("role")
	if roleRaw == nil {
		ac.logger.Warn("role not found in context")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	role, ok := roleRaw.(string)
	if !ok {
		ac.logger.Warn("role has invalid type")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	newTokenPair, err := ac.authService.TokenService.RefreshTokenPair(req.RefreshToken, userID, email, role)
	if err != nil {
		if errors.Is(err, domain_errors.ErrInvalidToken) || errors.Is(err, domain_errors.ErrRefreshTokenExpired) {
			ac.logger.Info("refresh token failed: invalid or expired token", slog.Any("user_id", userID))
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		ac.logger.Error("refresh token error", slog.Any("error", err), slog.Any("user_id", userID))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	ac.logger.Info("token refreshed", slog.Any("user_id", userID))

	return c.Status(http.StatusOK).JSON(auth_dto.AuthResponse{
		AccessToken:  newTokenPair.AccessToken,
		RefreshToken: newTokenPair.RefreshToken,
		ExpiresAt:    newTokenPair.ExpiresAt.String(),
	})
}

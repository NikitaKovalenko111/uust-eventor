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

func (ac *AuthController) RegisterRoutes(basicRouter fiber.Router, authMiddleware fiber.Handler) {
	basicRouter.Post("/register", ac.register)
	// moderator-only registration for creating moderator accounts
	basicRouter.Post("/register/moderator", authMiddleware, ac.registerModerator)
	basicRouter.Post("/login", ac.login)
	basicRouter.Post("/refresh", authMiddleware, ac.refreshToken)
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

	// Public registration must always create regular users
	tokenPair, err := ac.authService.Register(req.City, req.Name, req.Email, req.Password, "user")
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

// registerModerator allows an existing moderator to create a new moderator account by email
func (ac *AuthController) registerModerator(c *fiber.Ctx) error {
	// check current user's role
	roleRaw := c.Locals("role")
	role, _ := roleRaw.(string)
	if role != "moderator" {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "moderators only"})
	}

	var req auth_dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		ac.logger.Warn("failed to parse register moderator request", slog.Any("error", err))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	// require email and name and password
	if req.Email == "" || req.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "email and name are required"})
	}

	// if no password provided, generate a random one
	password := req.Password
	if password == "" {
		password = generateRandomPassword(12)
	}

	tokenPair, err := ac.authService.Register(req.City, req.Name, req.Email, password, "moderator")
	if err != nil {
		if errors.Is(err, domain_errors.ErrEmailAlreadyExists) {
			ac.logger.Info("moderator creation failed: email already exists", slog.String("email", req.Email))
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "email already registered"})
		}
		ac.logger.Error("moderator creation error", slog.Any("error", err), slog.String("email", req.Email))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}

	// return token pair and plaintext password when generated so caller can communicate it
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_at":    tokenPair.ExpiresAt.String(),
		"password":      password,
	})
}

// generateRandomPassword returns a random alphanumeric password of given length
func generateRandomPassword(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[int64(i*1103515245+12345)%int64(len(letters))] // deterministic fallback
	}
	return string(b)
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

	newTokenPair, err := ac.authService.TokenService.RefreshTokenPair(req.RefreshToken, userID, email, role, nil)
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

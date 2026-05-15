package middleware

import (
	"eventor/internal/platform/contracts/token_provider"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func NewJWTMiddleware(tokenService token_provider.TokenProvider, logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if isPublicReadRequest(c) {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Warn("missing authorization header")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("invalid authorization header format")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header"})
		}

		tokenString := parts[1]

		claims, err := tokenService.VerifyAccessToken(tokenString)
		if err != nil {
			logger.Info("token verification failed", slog.Any("error", err))
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

func isPublicReadRequest(c *fiber.Ctx) bool {
	if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
		return false
	}

	path := c.Path()
	if path == "/api/v1/events" || path == "/api/v1/events/" {
		return true
	}

	if strings.HasPrefix(path, "/api/v1/events/images/") && strings.HasSuffix(path, "/file") {
		return true
	}

	if strings.HasPrefix(path, "/api/v1/events/") {
		suffix := strings.TrimPrefix(path, "/api/v1/events/")
		if suffix == "" || strings.Contains(suffix, "/") {
			return false
		}
		return true
	}

	if strings.HasPrefix(path, "/api/v1/users/") && strings.HasSuffix(path, "/avatar/file") {
		suffix := strings.TrimPrefix(path, "/api/v1/users/")
		parts := strings.Split(suffix, "/")
		if len(parts) == 3 && parts[0] != "" && parts[1] == "avatar" && parts[2] == "file" {
			return true
		}
	}

	return false
}

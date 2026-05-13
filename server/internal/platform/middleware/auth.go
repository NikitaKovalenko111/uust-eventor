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

		return c.Next()
	}
}

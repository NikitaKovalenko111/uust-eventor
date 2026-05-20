package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

func NewLogger(logger *slog.Logger) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		logger.Info("NEW REQUEST", slog.String("PATH", string(c.Path())))

		return c.Next()
	}
}

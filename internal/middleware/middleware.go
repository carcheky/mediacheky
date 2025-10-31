package middleware

import (
	"time"

	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Logger middleware logs HTTP requests
func Logger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Process request
		err := c.Next()

		// Log request
		log.Info("HTTP request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", c.IP(),
			"request_id", c.Locals("requestid"),
		)

		return err
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for existing request ID in header
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context
		c.Locals("requestid", requestID)

		// Add to response header
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

// ErrorHandler is a custom error handler
func ErrorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		log.Error("Request error",
			"error", err,
			"path", c.Path(),
			"method", c.Method(),
			"status", code,
			"request_id", c.Locals("requestid"),
		)

		return c.Status(code).JSON(fiber.Map{
			"error":      err.Error(),
			"request_id": c.Locals("requestid"),
		})
	}
}

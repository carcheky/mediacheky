package middleware

import (
	"regexp"
	"strings"
	"time"

	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Compiled regex patterns for validation (package-level to avoid repeated compilation)
var (
	validServiceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	validConfigKeyRegex   = regexp.MustCompile(`^[A-Z0-9_]+$`)
)

// Logger middleware logs HTTP requests
func Logger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Process request
		err := c.Next()

		// Determine log level based on path and status
		// Reduce noise for frequently polled endpoints
		path := c.Path()
		status := c.Response().StatusCode()
		duration := time.Since(start).Milliseconds()

		// Use DEBUG level for:
		// - Successful requests to /api/services (polled frequently by multiple components)
		// - Successful requests to /api/services/:name GET (individual service status polling)
		// - Successful requests to /api/stats (polled every 30s)
		// - Successful requests to /status (health checks)
		if status >= 200 && status < 300 {
			if path == "/api/services" || strings.HasPrefix(path, "/api/services/") && c.Method() == "GET" || path == "/api/stats" || path == "/status" {
				log.Debug("HTTP request",
					"method", c.Method(),
					"path", path,
					"status", status,
					"duration_ms", duration,
					"ip", c.IP(),
					"request_id", c.Locals("requestid"),
				)
				return err
			}
		}

		// Log everything else as INFO
		log.Info("HTTP request",
			"method", c.Method(),
			"path", path,
			"status", status,
			"duration_ms", duration,
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

		// Only log 404s as debug, not errors (endpoints may not exist yet during setup)
		if code == fiber.StatusNotFound {
			log.Debug("Route not found",
				"path", c.Path(),
				"method", c.Method(),
				"status", code,
				"request_id", c.Locals("requestid"),
			)
		} else {
			log.Error("Request error",
				"error", err,
				"path", c.Path(),
				"method", c.Method(),
				"status", code,
				"request_id", c.Locals("requestid"),
			)
		}

		return c.Status(code).JSON(fiber.Map{
			"error":      err.Error(),
			"request_id": c.Locals("requestid"),
		})
	}
}

// ValidateServiceName validates service name parameter
// Service names should be alphanumeric with hyphens and underscores
func ValidateServiceName() fiber.Handler {
	return func(c *fiber.Ctx) error {
		serviceName := c.Params("name")
		if serviceName == "" {
			return c.Next() // No name parameter, skip validation
		}

		if !validServiceNameRegex.MatchString(serviceName) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid service name. Only alphanumeric characters, hyphens, and underscores are allowed",
			})
		}

		// Limit length to prevent abuse
		if len(serviceName) > 50 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Service name too long. Maximum 50 characters",
			})
		}

		return c.Next()
	}
}

// ValidateConfigKey validates configuration key parameter
func ValidateConfigKey() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Params("key")
		if key == "" {
			return c.Next() // No key parameter, skip validation
		}

		if !validConfigKeyRegex.MatchString(key) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid configuration key. Only uppercase alphanumeric characters and underscores are allowed",
			})
		}

		// Limit length
		if len(key) > 50 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Configuration key too long. Maximum 50 characters",
			})
		}

		return c.Next()
	}
}

// CORS configures CORS middleware with appropriate settings
func CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set CORS headers
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Set("Access-Control-Expose-Headers", "X-Request-ID")
		c.Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

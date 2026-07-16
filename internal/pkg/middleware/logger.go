package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

// LoggerConfig holds optional configuration for the request logger middleware.
type LoggerConfig struct {
	SkipPaths []string
}

// NewLogger returns a middleware that logs every incoming HTTP request.
// Completion logging is skipped when the handler returns an error — the
// centralized ErrorHandler takes care of error logging.
func NewLogger(log *logrus.Logger, cfg ...LoggerConfig) fiber.Handler {
	skip := make(map[string]bool)
	if len(cfg) > 0 {
		for _, p := range cfg[0].SkipPaths {
			skip[p] = true
		}
	}

	return func(c fiber.Ctx) error {
		if skip[string(c.Path())] {
			return c.Next()
		}

		start := time.Now()
		rid := c.Locals("request_id")
		if rid == nil {
			rid = ""
		}

		log.WithFields(logrus.Fields{
			"request_id": rid,
			"method":     c.Method(),
			"path":       c.Path(),
			"ip":         c.IP(),
		}).Info("incoming request")

		err := c.Next()

		// If the handler returned an error, skip completion logging —
		// the centralized ErrorHandler will log it.
		if err != nil {
			return err
		}

		log.WithFields(logrus.Fields{
			"request_id": rid,
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     c.Response().StatusCode(),
			"latency_ms": time.Since(start).Milliseconds(),
			"ip":         c.IP(),
		}).Info("request completed")

		return nil
	}
}

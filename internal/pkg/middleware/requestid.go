package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestIDConfig holds optional configuration for the request ID middleware.
type RequestIDConfig struct {
	Header string
	Locals string
}

var defaultRequestIDConfig = RequestIDConfig{
	Header: fiber.HeaderXRequestID,
	Locals: "request_id",
}

// NewRequestID returns a middleware that injects a unique request ID into
// the context and response header. If the client sends an X-Request-ID header,
// that value is reused; otherwise a new UUID is generated.
func NewRequestID(cfg ...RequestIDConfig) fiber.Handler {
	config := defaultRequestIDConfig
	if len(cfg) > 0 {
		if cfg[0].Header != "" {
			config.Header = cfg[0].Header
		}
		if cfg[0].Locals != "" {
			config.Locals = cfg[0].Locals
		}
	}

	return func(c fiber.Ctx) error {
		rid := c.Get(config.Header)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set(config.Header, rid)
		c.Locals(config.Locals, rid)
		return c.Next()
	}
}

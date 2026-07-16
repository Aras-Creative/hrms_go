package middleware

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	apperrors "hrms/internal/pkg/apperror"
	response "hrms/internal/pkg/api"
)

// NewRecover returns a middleware that catches panics, logs them with a
// stack trace, and returns a structured 500 JSON response.
func NewRecover(log *logrus.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				rid, _ := c.Locals("request_id").(string)

				log.WithFields(logrus.Fields{
					"request_id": rid,
					"path":       c.Path(),
					"method":     c.Method(),
					"panic":      r,
					"stack":      string(debug.Stack()),
				}).Error("panic recovered")

				_ = c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse{
					Success: false,
					Error: response.ErrorInfo{
						Code:    apperrors.ErrInternal.Code,
						Message: apperrors.ErrInternal.Message,
					},
				})
			}
		}()
		return c.Next()
	}
}

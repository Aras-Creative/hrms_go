package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	apperrors "hrms/internal/pkg/apperror"
	response "hrms/internal/pkg/api"
)

// NewErrorHandler returns a fiber.ErrorHandler that converts any error
// propagated by route handlers into a structured JSON response.
func NewErrorHandler(log *logrus.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		rid, _ := c.Locals("request_id").(string)

		// 1. Fiber-specific errors (route not found, 405, etc.)
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			log.WithFields(logrus.Fields{
				"request_id": rid,
				"status":     fiberErr.Code,
				"path":       c.Path(),
				"method":     c.Method(),
			}).Warn("fiber error: ", fiberErr.Message)

			return c.Status(fiberErr.Code).JSON(response.ErrorResponse{
				Success: false,
				Error: response.ErrorInfo{
					Code:    http.StatusText(fiberErr.Code),
					Message: fiberErr.Message,
				},
			})
		}

		// 2. Struct validation errors from go-playground/validator
		var validationErr validator.ValidationErrors
		if errors.As(err, &validationErr) {
			var msgs []string
			for _, e := range validationErr {
				msgs = append(msgs, e.Field()+": "+e.Tag()+" ("+e.Param()+")")
			}
			msg := strings.Join(msgs, "; ")

			log.WithFields(logrus.Fields{
				"request_id": rid,
				"path":       c.Path(),
				"method":     c.Method(),
				"validation": msg,
			}).Warn("validation error")

			return c.Status(fiber.StatusUnprocessableEntity).JSON(response.ErrorResponse{
				Success: false,
				Error: response.ErrorInfo{
					Code:    "VALIDATION_ERROR",
					Message: msg,
				},
			})
		}

		// 3. Our domain errors
		var domainErr *apperrors.DomainError
		if errors.As(err, &domainErr) {
			log.WithFields(logrus.Fields{
				"request_id": rid,
				"code":       domainErr.Code,
				"status":     domainErr.HTTPStatus,
				"path":       c.Path(),
				"method":     c.Method(),
			}).Warn("domain error: ", domainErr.Message)

			return response.Error(c, domainErr)
		}

		// 4. Unknown — log full details and return a generic 500
		log.WithFields(logrus.Fields{
			"request_id": rid,
			"path":       c.Path(),
			"method":     c.Method(),
			"error":      err.Error(),
		}).Error("unhandled error")

		return response.Error(c, apperrors.ErrInternal)
	}
}

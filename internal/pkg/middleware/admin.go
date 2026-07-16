package middleware

import (
	"github.com/gofiber/fiber/v3"
)

func NewAdminMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "admin access required",
				},
			})
		}
		return c.Next()
	}
}

package middleware

import (
	"github.com/gofiber/fiber/v3"

	"hrms/internal/pkg/iputil"
)

// NewIPWhitelistMiddleware restricts access to clients whose IP is covered by
// the CIDR allow-list loaded from settings (see iputil.SetAllowedCIDRs). An
// empty allow-list disables the restriction.
func NewIPWhitelistMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if !iputil.IsAllowed(c.IP()) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "IP_NOT_ALLOWED",
					"message": "your IP address is not allowed to punch",
				},
			})
		}
		return c.Next()
	}
}

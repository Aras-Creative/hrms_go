package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"hrms/internal/pkg/iputil"
)

// EmployeeResolver maps an authenticated user to their employee record so the
// middleware can apply per-employee exceptions.
type EmployeeResolver interface {
	FindByUserID(ctx context.Context, userID string) (employeeID, employeeName string, err error)
}

// NewIPWhitelistMiddleware restricts access to clients whose IP is covered by
// the CIDR allow-list loaded from settings (see iputil.SetAllowedCIDRs). An
// empty allow-list disables the restriction.
//
// Employees listed in the bypass list (see iputil.SetBypassEmployeeIDs) are
// exempt: when their IP is not allowed they are still granted access, provided
// the authenticated user can be resolved to a bypassed employee.
func NewIPWhitelistMiddleware(resolver EmployeeResolver) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		if ipAllowed(c.RequestCtx(), c.IP(), userID, resolver) {
			return c.Next()
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "IP_NOT_ALLOWED",
				"message": "your IP address is not allowed to punch",
			},
		})
	}
}

// ipAllowed reports whether a request from ip by userID may pass the allow-list.
// A bypassed employee is allowed regardless of their IP.
func ipAllowed(ctx context.Context, ip, userID string, resolver EmployeeResolver) bool {
	if iputil.IsAllowed(ip) {
		return true
	}
	if resolver == nil || userID == "" {
		return false
	}
	employeeID, _, err := resolver.FindByUserID(ctx, userID)
	return err == nil && iputil.IsEmployeeBypassed(employeeID)
}

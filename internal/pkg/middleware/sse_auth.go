package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	hrmsJwt "hrms/internal/pkg/jwt"
)

// NewSSEAuthMiddleware creates middleware that checks access_token cookie
// first, then falls back to ?token= query param (for EventSource which
// cannot set custom headers), and finally Authorization header.
func NewSSEAuthMiddleware(jwtSvc *hrmsJwt.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenStr := c.Cookies("access_token")
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}
		if tokenStr == "" {
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "missing access token",
				},
			})
		}

		claims, err := jwtSvc.ValidateAccessToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "invalid or expired access token",
				},
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

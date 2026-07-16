package delivery

import "github.com/gofiber/fiber/v3"

func (h *AuthHandler) RegisterRoutes(r fiber.Router) {
	auth := r.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/admin/login", h.LoginAdmin)
	auth.Post("/refresh", h.RefreshToken)
	auth.Get("/me", h.authMw, h.GetMe)
	auth.Post("/logout", h.authMw, h.Logout)
	auth.Post("/user/challenge", h.RequestChallenge)
	auth.Post("/user/login", h.LoginUser)
	auth.Delete("/devices/:userID", h.authMw, h.RevokeDevice)
	auth.Patch("/me/name", h.authMw, h.ChangeName)
}

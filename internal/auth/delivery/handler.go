package delivery

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"

	authAdapter "hrms/internal/auth/adapter"
	"hrms/internal/auth/models"
	"hrms/internal/auth/usecase"
	response "hrms/internal/pkg/api"
	errors "hrms/internal/pkg/apperror"
)

type Notifier interface {
	Notify(ctx context.Context, userID, ntype, title, body, resource, resourceID string) error
}

const notifTypeSecurity = "security"

type AuthHandler struct {
	uc              *usecase.AuthUsecase
	userUc          *usecase.UserUsecase
	authMw          fiber.Handler
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	secureCookies   bool
	auditLogger     *authAdapter.AuditLogger
	notifUC         Notifier
}

func NewAuthHandler(uc *usecase.AuthUsecase, userUc *usecase.UserUsecase, authMw fiber.Handler, accessTokenTTL, refreshTokenTTL time.Duration, secureCookies bool, auditLogger *authAdapter.AuditLogger, notifUC Notifier) *AuthHandler {
	return &AuthHandler{
		uc:              uc,
		userUc:          userUc,
		authMw:          authMw,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		secureCookies:   secureCookies,
		auditLogger:     auditLogger,
		notifUC:         notifUC,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var request RegisterRequest

	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	user, err := h.uc.Register(c.RequestCtx(), request.ToRegisterInput())
	if err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		uid, _ := c.Locals("user_id").(string)
		if uid == "" {
			uid = user.ID
		}
		h.auditLogger.Log(c.RequestCtx(), uid, "user", user.ID, "",
			authAdapter.ActionRegister, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"username": user.Username, "role": user.Role},
		)
	}

	return response.Created(c, NewRegisterResponse(user))
}

func (h *AuthHandler) LoginAdmin(c fiber.Ctx) error {
	var request LoginAdminRequest

	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	result, err := h.uc.LoginAdmin(c.RequestCtx(), request.ToLoginAdminInput())
	if err != nil {
		return response.Error(c, err)
	}

	h.setTokenCookies(c, result.AccessToken, result.RefreshToken)

	return response.OK(c, NewLoginAdminResponse(result))
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	result, err := h.uc.RefreshToken(c.RequestCtx(), models.RefreshTokenInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		h.clearTokenCookies(c)
		return response.Error(c, err)
	}

	h.setTokenCookies(c, result.AccessToken, result.RefreshToken)

	return response.OK(c, NewLoginAdminResponse(result))
}

func (h *AuthHandler) GetMe(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.ErrUnauthorized)
	}

	user, err := h.uc.GetMe(c.RequestCtx(), userID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, NewRegisterResponse(user))
}

func (h *AuthHandler) RevokeDevice(c fiber.Ctx) error {
	role, ok := c.Locals("role").(string)
	if !ok || (role != "super" && role != "admin") {
		return response.Error(c, errors.NewForbidden("admin access required"))
	}

	userID := c.Params("userID")
	if userID == "" {
		return response.Error(c, errors.NewInvalidInput("user_id is required"))
	}

	if err := h.uc.RevokeDevice(c.RequestCtx(), userID); err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID, _ := c.Locals("user_id").(string)
		h.auditLogger.Log(c.RequestCtx(), actorID, "user", userID, "",
			authAdapter.ActionDeviceRevoke, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"target_user_id": userID},
		)
	}

	if h.notifUC != nil {
		h.notifUC.Notify(c.RequestCtx(), userID, notifTypeSecurity,
			"Perangkat Dicabut",
			"Akses perangkat Anda telah dicabut oleh admin",
			"user", userID,
		)
	}

	return response.OK(c, fiber.Map{"message": "device revoked successfully"})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.ErrUnauthorized)
	}

	if err := h.uc.Logout(c.RequestCtx(), userID); err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		h.auditLogger.Log(c.RequestCtx(), userID, "user", userID, "",
			authAdapter.ActionLogout, c.IP(), string(c.RequestCtx().UserAgent()),
			nil,
		)
	}

	h.clearTokenCookies(c)

	return response.OK(c, nil)
}

func (h *AuthHandler) RequestChallenge(c fiber.Ctx) error {
	var request RequestChallengeRequest

	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	challenge, rawChallenge, err := h.uc.RequestChallenge(c.RequestCtx(), request.Username)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, ChallengeResponse{
		ChallengeID: challenge.ID,
		Challenge:   rawChallenge,
		ExpiresAt:   challenge.ExpiresAt,
	})
}

func (h *AuthHandler) LoginUser(c fiber.Ctx) error {
	var request UserLoginRequest

	if err := c.Bind().Body(&request); err != nil {
		return err
	}

	result, err := h.uc.LoginUser(c.RequestCtx(), models.UserLoginInput{
		Username:    request.Username,
		ChallengeID: request.ChallengeID,
		Challenge:   request.Challenge,
		Signature:   request.Signature,
		PublicKey:   request.PublicKey,
		DeviceName:  request.Device.Name,
		Platform:    request.Device.Platform,
	})
	if err != nil {
		return response.Error(c, err)
	}

	h.setTokenCookies(c, result.AccessToken, result.RefreshToken)

	return response.OK(c, NewLoginAdminResponse(result))
}

func (h *AuthHandler) ChangeName(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, errors.ErrUnauthorized)
	}

	var req ChangeNameRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	if err := h.userUc.ChangeName(c.RequestCtx(), &usecase.ChangeNameInput{
		UserID:   userID,
		FullName: req.FullName,
		Username: req.Username,
	}); err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, nil)
}

func (h *AuthHandler) ChangePassword(c fiber.Ctx) error {
	role, ok := c.Locals("role").(string)
	if !ok || (role != "super" && role != "admin") {
		return response.Error(c, errors.NewForbidden("admin access required"))
	}

	var req ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	if err := h.uc.ChangePassword(c.RequestCtx(), req.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		return response.Error(c, err)
	}

	if h.auditLogger != nil {
		actorID, _ := c.Locals("user_id").(string)
		h.auditLogger.Log(c.RequestCtx(), actorID, "user", req.UserID, "",
			authAdapter.ActionPasswordChange, c.IP(), string(c.RequestCtx().UserAgent()),
			map[string]any{"target_user_id": req.UserID},
		)
	}

	if h.notifUC != nil {
		h.notifUC.Notify(c.RequestCtx(), req.UserID, notifTypeSecurity,
			"Kata Sandi Diubah",
			"Kata sandi akun Anda telah diubah oleh admin",
			"user", req.UserID,
		)
	}

	return response.OK(c, fiber.Map{"message": "password changed successfully"})
}

func (h *AuthHandler) setTokenCookies(c fiber.Ctx, accessToken, refreshToken string) {
	accessMaxAge := int(h.accessTokenTTL.Seconds())
	refreshMaxAge := int(h.refreshTokenTTL.Seconds())

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   accessMaxAge,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   refreshMaxAge,
	})
}

func (h *AuthHandler) clearTokenCookies(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   h.secureCookies,
		SameSite: "Lax",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

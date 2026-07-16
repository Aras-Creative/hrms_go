package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"hrms/internal/auth/adapter"
	"hrms/internal/auth/entity"
	"hrms/internal/auth/models"
	"hrms/internal/auth/repository"
	errors "hrms/internal/pkg/apperror"
)

type AuthUsecase struct {
	userRepo      repository.UserRepository
	sessionRepo   repository.SessionRepository
	challengeRepo repository.ChallengeRepository
	deviceRepo    repository.DeviceRepository
	hasher        Hasher
	challengeHash ChallengeHasher
	token         TokenGenerator
	accessTTL     time.Duration
	refreshTTL    time.Duration
	challengeTTL  time.Duration
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	challengeRepo repository.ChallengeRepository,
	deviceRepo repository.DeviceRepository,
	hasher Hasher,
	challengeHash ChallengeHasher,
	token TokenGenerator,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	challengeTTL time.Duration,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		challengeRepo: challengeRepo,
		deviceRepo:    deviceRepo,
		hasher:        hasher,
		challengeHash: challengeHash,
		token:         token,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		challengeTTL:  challengeTTL,
	}
}

func (uc *AuthUsecase) Register(ctx context.Context, input *models.RegisterUserInput) (*entity.User, error) {
	exists, err := uc.userRepo.Exists(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("failed to register user: %w", errors.NewAlreadyExists("user already exists"))
	}

	hashedPassword, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role, err := entity.ParseRole(input.Role)
	if err != nil {
		return nil, err
	}

	newUser := entity.NewUserWithRole(input.Username, hashedPassword, input.FullName, role)

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}

func (uc *AuthUsecase) LoginAdmin(ctx context.Context, input models.LoginAdminInput) (*models.LoginResult, error) {
	_ = uc.sessionRepo.DeleteExpired(ctx)

	user, err := uc.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if user == nil {
		return nil, errors.NewUnauthorized("invalid credentials")
	}

	if !user.IsAdmin() {
		return nil, errors.NewForbidden("user does not have admin privileges")
	}

	if err := uc.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		return nil, errors.NewUnauthorized("invalid credentials")
	}

	sessions, err := uc.sessionRepo.FindActiveByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing sessions: %w", err)
	}
	for _, s := range sessions {
		s.Revoke()
		if err := uc.sessionRepo.UpdateSession(ctx, s); err != nil {
			return nil, fmt.Errorf("failed to revoke session: %w", err)
		}
	}

	accessToken, err := uc.token.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.token.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	session := entity.NewSession(user.ID, "", refreshToken, time.Now().Add(uc.refreshTTL))
	if err := uc.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &models.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		Session:      session,
	}, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, input models.RefreshTokenInput) (*models.LoginResult, error) {
	if input.RefreshToken == "" {
		return nil, errors.NewUnauthorized("missing refresh token")
	}

	session, err := uc.sessionRepo.FindByRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	if session == nil {
		return nil, errors.NewUnauthorized("invalid or expired refresh token")
	}

	if session.IsExpired() {
		return nil, errors.NewUnauthorized("refresh token expired")
	}

	user, err := uc.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, errors.NewUnauthorized("user not found or inactive")
	}

	session.Revoke()
	if err := uc.sessionRepo.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to revoke old session: %w", err)
	}

	accessToken, err := uc.token.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := uc.token.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	newSession := entity.NewSession(user.ID, session.DeviceID, newRefreshToken, time.Now().Add(uc.refreshTTL))
	if err := uc.sessionRepo.CreateSession(ctx, newSession); err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	return &models.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
		Session:      newSession,
	}, nil
}

func (uc *AuthUsecase) GetMe(ctx context.Context, userID string) (*entity.User, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, errors.NewUnauthorized("user not found or inactive")
	}

	return user, nil
}

func (uc *AuthUsecase) RevokeDevice(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.NewNotFound("user not found")
	}

	if err := uc.deviceRepo.RevokeDeviceByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke device: %w", err)
	}

	sessions, err := uc.sessionRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find sessions: %w", err)
	}
	for _, s := range sessions {
		s.Revoke()
		if err := uc.sessionRepo.UpdateSession(ctx, s); err != nil {
			return fmt.Errorf("failed to revoke session: %w", err)
		}
	}

	return nil
}

func (uc *AuthUsecase) GetDeviceByUserID(ctx context.Context, userID string) (*models.DeviceInfo, error) {
	device, err := uc.deviceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find device: %w", err)
	}
	if device == nil {
		return nil, nil
	}
	return &models.DeviceInfo{
		ID:         device.ID,
		Platform:   device.Platform,
		Name:       device.UserAgent,
		IsActive:   device.IsActive,
		LastUsedAt: device.LastUsedAt,
		CreatedAt:  device.CreatedAt,
	}, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, userID string) error {
	sessions, err := uc.sessionRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find sessions: %w", err)
	}
	for _, s := range sessions {
		s.Revoke()
		if err := uc.sessionRepo.UpdateSession(ctx, s); err != nil {
			return fmt.Errorf("failed to revoke session: %w", err)
		}
	}
	return nil
}

func (uc *AuthUsecase) RequestChallenge(ctx context.Context, username string) (*entity.Challenge, string, error) {
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, "", errors.NewUnauthorized("user not found or inactive")
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, "", fmt.Errorf("failed to generate challenge: %w", err)
	}
	rawChallenge := hex.EncodeToString(b)

	hash, err := uc.challengeHash.HashChallenge(rawChallenge)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash challenge: %w", err)
	}

	challenge := entity.NewChallenge(user.ID, hash, uc.challengeTTL)
	if err := uc.challengeRepo.CreateChallenge(ctx, challenge); err != nil {
		return nil, "", fmt.Errorf("failed to save challenge: %w", err)
	}

	return challenge, rawChallenge, nil
}

func (uc *AuthUsecase) LoginUser(ctx context.Context, input models.UserLoginInput) (*models.LoginResult, error) {
	_ = uc.sessionRepo.DeleteExpired(ctx)

	user, err := uc.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, errors.NewUnauthorized("invalid credentials")
	}
	if !user.IsUser() {
		return nil, errors.NewForbidden("only regular users can login with device")
	}

	// Verify challenge
	challenge, err := uc.challengeRepo.FindByID(ctx, input.ChallengeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find challenge: %w", err)
	}
	if challenge == nil {
		return nil, errors.NewUnauthorized("invalid credentials")
	}

	if err := challenge.ValidateFor(user.ID); err != nil {
		return nil, errors.NewUnauthorized("invalid credentials")
	}

	if err := uc.challengeHash.VerifyChallenge(challenge.ChallengeHash, input.Challenge); err != nil {
		return nil, errors.NewUnauthorized("invalid credentials")
	}

	// Verify signature against provided or stored public key
	pubKey, err := adapter.ParsePublicKey(input.PublicKey)
	if err != nil {
		return nil, errors.NewUnauthorized("invalid credentials")
	}
	if err := adapter.VerifySignature(pubKey, input.Challenge, input.Signature); err != nil {
		return nil, errors.NewUnauthorized("invalid credentials")
	}

	// Find or bind device
	device, err := uc.deviceRepo.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find device: %w", err)
	}

	if device == nil || !device.IsActiveDevice() {
		if device != nil && !device.IsActiveDevice() {
			// Device was revoked — revoke any remaining sessions
			sessions, err := uc.sessionRepo.FindActiveByUserID(ctx, user.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to find sessions: %w", err)
			}
			for _, s := range sessions {
				s.Revoke()
				if err := uc.sessionRepo.UpdateSession(ctx, s); err != nil {
					return nil, fmt.Errorf("failed to revoke session: %w", err)
				}
			}
		}
		device = entity.NewDevice(user.ID, input.PublicKey, input.Platform, input.DeviceName)
		if err := uc.deviceRepo.CreateDevice(ctx, device); err != nil {
			return nil, fmt.Errorf("failed to bind device: %w", err)
		}
	} else {
		storedPub, err := adapter.ParsePublicKey(device.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("invalid stored public key: %w", err)
		}
		if err := adapter.VerifySignature(storedPub, input.Challenge, input.Signature); err != nil {
			return nil, errors.NewUnauthorized("invalid credentials")
		}
	}

	// Atomic challenge use — prevents replay via TOCTOU
	if err := uc.challengeRepo.UpdateChallenge(ctx, challenge); err != nil {
		return nil, errors.NewUnauthorized("invalid credentials")
	}

	// One-device policy — check for conflict with different device
	activeSessions, err := uc.sessionRepo.FindActiveByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active sessions: %w", err)
	}
	if err := entity.CheckOneDevicePolicy(activeSessions, device.ID); err != nil {
		return nil, errors.NewForbidden(err.Error())
	}
	// Persist any revocations made by the policy check
	for _, s := range activeSessions {
		if s.BelongsToDevice(device.ID) && !s.IsActive {
			if err := uc.sessionRepo.UpdateSession(ctx, s); err != nil {
				return nil, fmt.Errorf("failed to revoke old session: %w", err)
			}
		}
	}

	// Update device last_used
	device.Touch()
	if err := uc.deviceRepo.UpdateDevice(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	// Generate tokens
	accessToken, err := uc.token.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.token.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	session := entity.NewSession(user.ID, device.ID, refreshToken, time.Now().Add(uc.refreshTTL))
	if err := uc.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &models.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		Session:      session,
	}, nil
}

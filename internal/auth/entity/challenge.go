package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Challenge struct {
	ID            string
	UserID        string
	ChallengeHash string
	ExpiresAt     time.Time
	IsUsed        bool
	CreatedAt     time.Time
}

func NewChallenge(userID, challengeHash string, ttl time.Duration) *Challenge {
	return &Challenge{
		ID:            uuid.New().String(),
		UserID:        userID,
		ChallengeHash: challengeHash,
		ExpiresAt:     time.Now().Add(ttl),
		IsUsed:        false,
		CreatedAt:     time.Now(),
	}
}

func ReconstituteChallenge(
	id, userID, challengeHash string,
	expiresAt time.Time,
	isUsed bool,
	createdAt time.Time,
) *Challenge {
	return &Challenge{
		ID:            id,
		UserID:        userID,
		ChallengeHash: challengeHash,
		ExpiresAt:     expiresAt,
		IsUsed:        isUsed,
		CreatedAt:     createdAt,
	}
}

func (c *Challenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *Challenge) IsUsedChallenge() bool {
	return c.IsUsed
}

func (c *Challenge) Use() {
	c.IsUsed = true
}

// ValidateFor checks that the challenge is valid for the given user:
// not used, not expired, and belongs to that user.
func (c *Challenge) ValidateFor(userID string) error {
	if c.IsUsed {
		return fmt.Errorf("challenge already used")
	}
	if c.IsExpired() {
		return fmt.Errorf("challenge expired")
	}
	if c.UserID != userID {
		return fmt.Errorf("challenge does not belong to this user")
	}
	return nil
}

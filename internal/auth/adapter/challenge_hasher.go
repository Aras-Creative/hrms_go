package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type SHA256ChallengeHasher struct{}

func NewSHA256ChallengeHasher() *SHA256ChallengeHasher {
	return &SHA256ChallengeHasher{}
}

func (h *SHA256ChallengeHasher) HashChallenge(challenge string) (string, error) {
	hash := sha256.Sum256([]byte(challenge))
	return hex.EncodeToString(hash[:]), nil
}

func (h *SHA256ChallengeHasher) VerifyChallenge(hash, challenge string) error {
	computed, err := h.HashChallenge(challenge)
	if err != nil {
		return err
	}
	if computed != hash {
		return fmt.Errorf("challenge mismatch")
	}
	return nil
}

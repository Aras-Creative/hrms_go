package adapter

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/big"
)

type ECDSASigner struct{}

func NewECDSASigner() *ECDSASigner {
	return &ECDSASigner{}
}

func (s *ECDSASigner) GenerateKey() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ecdsa key: %w", err)
	}
	return privateKey, nil
}

func ParsePublicKey(encoded string) (*ecdsa.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	pub, err := x509.ParsePKIXPublicKey(raw)
	if err == nil {
		ecdsaKey, ok := pub.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not an ECDSA public key")
		}
		return ecdsaKey, nil
	}

	if len(raw) == 64 {
		half := len(raw) / 2
		x := new(big.Int).SetBytes(raw[:half])
		y := new(big.Int).SetBytes(raw[half:])
		return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
	}

	return nil, fmt.Errorf("invalid public key format")
}

func VerifySignature(pub *ecdsa.PublicKey, challenge, signatureB64 string) error {
	sig, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return fmt.Errorf("invalid signature base64: %w", err)
	}

	hash := sha256.Sum256([]byte(challenge))

	if ecdsa.VerifyASN1(pub, hash[:], sig) {
		return nil
	}

	if len(sig) == 64 {
		r := new(big.Int).SetBytes(sig[:len(sig)/2])
		s := new(big.Int).SetBytes(sig[len(sig)/2:])
		if ecdsa.Verify(pub, hash[:], r, s) {
			return nil
		}
	}

	return fmt.Errorf("signature verification failed")
}

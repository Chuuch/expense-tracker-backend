package utils

import (
	"crypto/rand"
	"fmt"
)

func GenerateVerificationCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}

	n := 100000 + (uint32(b[0])|uint32(b[1])<<8|uint32(b[2])<<16|uint32(b[3])<<24)%900000
	return fmt.Sprintf("%06d", n), nil
}

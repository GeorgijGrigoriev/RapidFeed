package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GeneratePassword(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b), nil
}

func GenerateToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", errors.New("token length must be positive")
	}

	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

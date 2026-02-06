package db

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
)

func GetToken(token string) (models.Token, error) {
	var tokenInfo models.Token

	err := DB.QueryRow(
		`SELECT id, token, expires_at, permissions FROM token_storage WHERE token = ?`, token).Scan(&tokenInfo)

	if err != nil {
		slog.Error("failed to get token info", "token info", token)

		return tokenInfo, fmt.Errorf("failed to get token info: %w", err)
	}

	return tokenInfo, nil
}

func AddToken(userID, permission int, valid time.Duration) error {
	expiration := time.Now().Add(valid)

	token, err := generateToken(32)
	if err != nil {
		slog.Error("failed to generate token", "error", err)

		return fmt.Errorf("failed to generate token %w", err)
	}

	insertTokenQuery := `INSERT INTO token_storage (user_id, token, expires_at, permissions) VALUES (?, ?, ?, ?)`

	_, err = DB.Exec(insertTokenQuery, userID, token, expiration.Unix(), permission)
	if err != nil {
		slog.Error("failed to insert token info", "error", err)

		return fmt.Errorf("failed to insert token info: %w", err)
	}

	return nil
}

func RevokeToken(token string) error {
	_, err := DB.Exec(`DELETE FROM token_storage WHERE token = ?`, token)
	if err != nil {
		slog.Error("failed to delete token", "error", err)

		return fmt.Errorf("failed to delete token %w", err)
	}

	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateToken returns a random token with a-zA-Z0-9 pattern.
func generateToken(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("token length must be greater than zero")
	}

	token := make([]byte, length)

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		token[i] = charset[idx.Int64()]
	}

	return string(token), nil
}

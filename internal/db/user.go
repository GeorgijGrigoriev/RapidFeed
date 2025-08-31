package db

import (
	"fmt"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
)

func RegisterUser(username, password string) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hashing error: %w", err)
	}

	_, err = DB.Exec(`INSERT INTO users (username, password, role) VALUES (?, ?, 'user')`, username, hash)
	if err != nil {
		return fmt.Errorf("insert error: %w", err)
	}

	return nil
}

func ChangeUserPassword(userId int, password string) error {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hashing error: %w", err)
	}

	// Update password in database
	_, err = DB.Exec(`UPDATE users SET password = ? WHERE id = ?`, hashedPassword, userId)
	if err != nil {
		return fmt.Errorf("failed to change user password: %w", err)
	}

	return nil
}

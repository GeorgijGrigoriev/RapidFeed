package db

import (
	"fmt"
	"strconv"
	"strings"

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

func AddUser(username, password, role string) error {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = DB.Exec(`INSERT INTO users (username, password, role) VALUES (?, ?, ?)`, username, hashedPassword, role)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "duplicate key") {

			return fmt.Errorf("user already exists: %w", err)
		}

		return fmt.Errorf("failed to create new user: %w", err)
	}

	return nil
}

func BlockUser(userId string) error {
	blockUserId, err := strconv.Atoi(userId)
	if err != nil {
		return fmt.Errorf("invalid user id for block request: %w", err)
	}

	_, err = DB.Exec(`UPDATE users SET role = 'blocked' WHERE id = ?`, blockUserId)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	return nil
}

func UnblockUser(userId string) error {
	blockUserId, err := strconv.Atoi(userId)
	if err != nil {
		return fmt.Errorf("invalid user id for unblock request: %w", err)
	}

	_, err = DB.Exec(`UPDATE users SET role = 'user' WHERE id = ?`, blockUserId)
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	return nil
}

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
	"log/slog"
	"os"
)

var DB *sql.DB

var ErrTokenNotFound = errors.New("token not found")

func InitDB() {
	db, err := sql.Open("sqlite3", "./feeds.db")
	if err != nil {
		slog.Error("failed to initialize database connection", "error", err)

		os.Exit(1)
	}

	DB = db
}

func GetUserInfo(userID int) (models.User, error) {
	var user models.User

	err := DB.QueryRow(`SELECT id, username, role FROM users WHERE id = ?`, userID).Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {
		slog.Error("failed to get user info", "userID", userID)

		return user, fmt.Errorf("failed to get user info: %w", err)
	}

	return user, nil
}

func GetUserRole(userID int) (string, error) {
	var role string

	err := DB.QueryRow(`SELECT role FROM users WHERE id = ?`, userID).Scan(&role)
	if err != nil {
		slog.Error("failed to get user role", "userID", userID)

		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	return role, nil
}

func GetUserIDByToken(token string) (int, error) {
	var userID int

	err := DB.QueryRow(`SELECT user_id FROM user_tokens WHERE token = ?`, token).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrTokenNotFound
		}
		return 0, fmt.Errorf("failed to get user id by token: %w", err)
	}

	return userID, nil
}

func GetUserToken(userID int) (string, error) {
	var token string

	err := DB.QueryRow(`SELECT token FROM user_tokens WHERE user_id = ?`, userID).Scan(&token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrTokenNotFound
		}
		return "", fmt.Errorf("failed to get user token: %w", err)
	}

	return token, nil
}

func UpsertUserToken(userID int, token string) error {
	_, err := DB.Exec(`INSERT INTO user_tokens (user_id, token) VALUES (?, ?)
        ON CONFLICT(user_id) DO UPDATE SET token = excluded.token, created_at = CURRENT_TIMESTAMP`, userID, token)
	if err != nil {
		return fmt.Errorf("failed to upsert user token: %w", err)
	}

	return nil
}

func DeleteUserToken(userID int) error {
	_, err := DB.Exec(`DELETE FROM user_tokens WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user token: %w", err)
	}

	return nil
}

func GetUserFeeds(userID int) ([]models.UserFeed, error) {
	var userFeeds []models.UserFeed

	rows, err := DB.Query(`SELECT id, feed_url, title FROM user_feeds WHERE user_id = ?`, userID)
	if err != nil {
		slog.Error("failed to get user feeds", "userID", userID)

		return userFeeds, fmt.Errorf("failed to get user feeds: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var feed models.UserFeed
		err := rows.Scan(&feed.ID, &feed.FeedURL, &feed.Title)
		if err != nil {
			slog.Error("failed to scan user feed rows", "userID", userID)

			return userFeeds, fmt.Errorf("failed to scan user feed rows: %w", err)
		}
		userFeeds = append(userFeeds, feed)
	}

	return userFeeds, nil
}

func GetUserFeedUrls(userID int) ([]string, error) {
	var userFeeds []string

	rows, err := DB.Query(`SELECT feed_url FROM user_feeds WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to select user feeds, error: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var url string

		err := rows.Scan(&url)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user feeds, error: %w", err)
		}

		userFeeds = append(userFeeds, url)
	}

	return userFeeds, nil
}

func GetUsers() ([]models.User, error) {
	var users []models.User

	rows, err := DB.Query(`SELECT id, username, role FROM users`)
	if err != nil {
		slog.Error("failed to get users", "error", err)

		return users, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role)
		if err != nil {
			slog.Error("failed to scan users", "error", err)

			return users, fmt.Errorf("failed to scan users: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func GetUsersWithFeeds() ([]models.UserWithFeeds, error) {
	var usersWithFeeds []models.UserWithFeeds

	users, err := GetUsers()
	if err != nil {
		slog.Error("failed to get users", "error", err)

		return usersWithFeeds, fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range users {
		var userFeeds []models.UserFeed

		rows, err := DB.Query(`SELECT id, feed_url, title FROM user_feeds WHERE user_id = ?`, user.ID)
		if err != nil {
			slog.Error("failed to get user feeds", "user", user)

			return usersWithFeeds, fmt.Errorf("failed to get user feeds: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var feed models.UserFeed
			err := rows.Scan(&feed.ID, &feed.FeedURL, &feed.Title)
			if err != nil {
				slog.Error("failed to scan user feed rows", "user", user)

				return usersWithFeeds, fmt.Errorf("failed to scan user feed rows: %w", err)
			}
			userFeeds = append(userFeeds, feed)
		}

		usersWithFeeds = append(usersWithFeeds, models.UserWithFeeds{
			User:      user,
			UserFeeds: userFeeds,
		})
	}

	return usersWithFeeds, nil
}

package db

import (
	"log"
	"log/slog"
	"os"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
)

const defaultPasswordLength = 14

func InitSchema() {
	createTableQuery := `CREATE TABLE IF NOT EXISTS feeds (
        "id" INTEGER PRIMARY KEY AUTOINCREMENT,
        "title" TEXT,
        "link" TEXT,
        "date" TIMESTAMP,
        "source" TEXT,
		"description" TEXT,
        "feed_url" TEXT
    )`

	_, err := DB.Exec(createTableQuery)
	if err != nil {
		slog.Error("failed to create feeds table", "error", err)

		os.Exit(1)
	}

	createUsersTableQuery := `CREATE TABLE IF NOT EXISTS users (
        "id" INTEGER PRIMARY KEY AUTOINCREMENT,
        "username" TEXT UNIQUE,
        "password" TEXT,
        "role" TEXT CHECK( role IN ('user', 'admin', 'blocked') )
    )`

	_, err = DB.Exec(createUsersTableQuery)
	if err != nil {
		slog.Error("failed to create users table", "error", err)

		os.Exit(1)
	}

	createUserFeedsTableQuery := `CREATE TABLE IF NOT EXISTS user_feeds (
        "id" INTEGER PRIMARY KEY AUTOINCREMENT,
        "user_id" INTEGER,
        "feed_url" TEXT,
        "title" TEXT,
        "category" TEXT,
        FOREIGN KEY("user_id") REFERENCES users("id")
    )`

	_, err = DB.Exec(createUserFeedsTableQuery)
	if err != nil {
		slog.Error("failed to create user_feeds table", "error", err)

		os.Exit(1)
	}

	createFeedLinkDateIndex := `CREATE INDEX IF NOT EXISTS idx_feeds_link_feedurl_date ON feeds (link, feed_url, date);`

	_, err = DB.Exec(createFeedLinkDateIndex)
	if err != nil {
		slog.Error("failed to create feed_link_date index", "error", err)

		os.Exit(1)
	}

	_, err = DB.Exec(`
              CREATE TABLE IF NOT EXISTS user_refresh_settings (
                      user_id INTEGER PRIMARY KEY,
                      interval_minutes INTEGER DEFAULT 60,
					  last_update_ts STRING,
                      FOREIGN KEY (user_id) REFERENCES users(id)
              )`)
	if err != nil {
		slog.Error("failed to create user_refresh_settings table", "error", err)

		os.Exit(1)
	}

	createTokenTable := `CREATE TABLE IF NOT EXISTS token_storage (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"user_id" INTEGER,
		"token" TEXT NOT NULL,
		"expires_at" INTEGER,
		"permissions" INTEGER
	)`

	_, err = DB.Exec(createTokenTable)
	if err != nil {
		slog.Error("failed to create token_storage table", "error", err)
		os.Exit(1)
	}
}

func CreateDefaultAdmin() {
	var adminExists bool

	err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = 'admin')`).Scan(&adminExists)
	if err != nil {
		log.Fatal(err)
	}

	adminPass, err := auth.GeneratePassword(defaultPasswordLength)
	if err != nil {
		slog.Error("failed to generate password for admin", "error", err)

		os.Exit(1)
	}

	encryptedPass, err := auth.HashPassword(utils.GetStringEnv("ADMIN_PASSWORD", adminPass))
	if err != nil {
		slog.Error("failed to hash default admin password", "error", err)

		os.Exit(1)
	}

	if !adminExists {
		insertAdminQuery := `INSERT INTO users (username, password, role) VALUES ('admin', ? , 'admin')`

		_, err = DB.Exec(insertAdminQuery, encryptedPass)
		if err != nil {
			slog.Error("failed to create default admin", "error", err)

			os.Exit(1)
		}

		slog.Info("!!!!!!!!!!")
		slog.Info("created default admin", "password", adminPass)
		slog.Info("this password shown only this time, keep it in safe place")
		slog.Info("!!!!!!!!!!")
	}
}

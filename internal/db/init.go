package db

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"log"
	"log/slog"
	"os"
)

func InitSchema() {
	createTableQuery := `CREATE TABLE IF NOT EXISTS feeds (
        "id" INTEGER PRIMARY KEY AUTOINCREMENT,
        "title" TEXT,
        "link" TEXT,
        "date" TEXT,
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
        "role" TEXT CHECK( role IN ('user', 'admin') )
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
}

func CreateDefaultAdmin() {
	var adminExists bool

	err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = 'admin')`).Scan(&adminExists)
	if err != nil {
		log.Fatal(err)
	}

	encryptedPass, err := auth.HashPassword(utils.GetStringEnv("ADMIN_PASSWORD", "admin"))
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
	}
}

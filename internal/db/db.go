package db

import (
	"database/sql"
	"log/slog"
	"os"
)

var DB *sql.DB

func InitDB() {
	db, err := sql.Open("sqlite3", "./feeds.db")
	if err != nil {
		slog.Error("failed to initialize database connection", "error", err)

		os.Exit(1)
	}

	DB = db
}

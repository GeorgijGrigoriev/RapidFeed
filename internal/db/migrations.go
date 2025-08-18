package db

import (
	"log/slog"
	"os"
	"strings"
)

// migrationAddNextUpdateTS - run migration on user_refresh_settings table for add additional column
func migrationAddNextUpdateTS() {
	_, err := DB.Exec(`ALTER TABLE user_refresh_settings ADD COLUMN next_update_ts TEXT`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		slog.Error("failed to add next_update_ts to user_refresh_settings", "error", err)

		os.Exit(1)
	}
}

// RunAllMigrations - running all possible migrations
func RunAllMigrations() {
	migrationAddNextUpdateTS()
}

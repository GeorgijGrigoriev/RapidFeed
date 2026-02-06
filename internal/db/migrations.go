package db

import (
	"log/slog"
	"os"
	"strings"
)

// migrationAddNextUpdateTS - run migration on user_refresh_settings table for add additional column.
func migrationAddNextUpdateTS() {
	_, err := DB.Exec(`ALTER TABLE user_refresh_settings ADD COLUMN next_update_ts TEXT`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		slog.Error("failed to add next_update_ts to user_refresh_settings", "error", err)

		os.Exit(1)
	}
}

// migrationBackfillUserFeedsCategory - ensure category column exists and has no NULL values.
func migrationBackfillUserFeedsCategory() {
	_, err := DB.Exec(`ALTER TABLE user_feeds ADD COLUMN category TEXT`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		slog.Error("failed to add category to user_feeds", "error", err)

		os.Exit(1)
	}

	_, err = DB.Exec(`UPDATE user_feeds SET category = '' WHERE category IS NULL`)
	if err != nil {
		slog.Error("failed to backfill NULL category values in user_feeds", "error", err)

		os.Exit(1)
	}
}

// RunAllMigrations - running all possible migrations.
func RunAllMigrations() {
	migrationAddNextUpdateTS()
	migrationBackfillUserFeedsCategory()
}

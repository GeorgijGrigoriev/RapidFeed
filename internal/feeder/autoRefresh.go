package feeder

import (
	"log/slog"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
)

// StartAutoRefresh starts the auto-refresh service for all users
func StartAutoRefresh() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for range ticker.C {
		slog.Info("updating user feeds check")
		refreshAllUsersFeeds()
	}
}

// refreshAllUsersFeeds refreshes feeds for all users based on their settings
func refreshAllUsersFeeds() {
	users, err := db.GetUsers()
	if err != nil {
		slog.Error("failed to get all users", "error", err)

		return
	}

	for _, user := range users {
		interval, err := db.GetUserRefreshInterval(user.ID)
		if err != nil {
			slog.Error("failed to get refresh interval for user", "user_id", user.ID, "error", err)

			continue
		}

		if interval == 0 {
			slog.Info("autofetch is disabled", "user id", user.ID, "user name", user.Username)

			continue
		}

		lastUpdateTS, err := db.GetLastUpdateTS(user.ID)
		if err != nil {
			slog.Error("failed to get last update TS", "error", err)

			continue
		}

		shouldUpdate := lastUpdateTS.IsZero() || time.Now().After(lastUpdateTS)

		if !shouldUpdate {
			continue
		}

		slog.Info("updating user feeds with autorefresh", "user id", user.ID, "user name", user.Username)

		feedUrls, err := db.GetUserFeedUrls(user.ID)
		if err != nil {
			slog.Error("failed to get user feeds", "user_id", user.ID, "error", err)

			continue
		}

		// Only fetch if user has feeds
		if len(feedUrls) > 0 {
			FetchAndSaveFeeds(feedUrls)
		}

		err = db.SetLastUpdateTS(user.ID, interval)
		if err != nil {
			slog.Error("failed to update last_update_ts", "error", err)
		}
	}
}

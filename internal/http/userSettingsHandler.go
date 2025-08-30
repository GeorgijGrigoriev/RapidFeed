package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
)

func userSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := checkSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	switch r.Method {
	case http.MethodGet:
		userFeeds, err := db.GetUserFeeds(userID)
		if err != nil {
			internalServerErrorHandler(w, r, err)

			return
		}

		user, err := db.GetUserInfoById(userID)
		if err != nil {
			internalServerErrorHandler(w, r, err)

			return
		}

		refreshInterval, err := db.GetUserRefreshInterval(user.ID)
		if err != nil {
			internalServerErrorHandler(w, r, err)

			return
		}

		tmpl := PrepareTemplate("internal/templates/base.html",
			"internal/templates/navbar.html",
			"internal/templates/settings.html")

		data := map[string]any{
			"UserFeeds":       userFeeds,
			"User":            user,
			"Title":           "Settings - RapidFeed",
			"RefreshInterval": refreshInterval,
		}

		tmpl.ExecuteTemplate(w, "base", data)
	case http.MethodPost:
		if r.FormValue("current_password") != "" && r.FormValue("new_password") != "" {
			currentPassword := r.FormValue("current_password")
			newPassword := r.FormValue("new_password")

			// Get user's current password hash from database
			user, err := db.GetUserInfoById(userID)
			if err != nil {
				internalServerErrorHandler(w, r, err)
				return
			}

			hash, err := db.GetUserHash(user.Username)
			if err != nil {
				slog.Error("failed to get user hash", "error", err)
				internalServerErrorHandler(w, r, fmt.Errorf("failed to verify current password"))
				return
			}

			// Verify current password
			err = auth.CheckPassword(hash, currentPassword)
			if err != nil {
				slog.Error("incorrect current password", "error", err)
				http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
				return
			}

			// Hash new password
			hashedPassword, err := auth.HashPassword(newPassword)
			if err != nil {
				slog.Error("failed to hash new password", "error", err)
				internalServerErrorHandler(w, r, fmt.Errorf("failed to hash new password"))
				return
			}

			// Update password in database
			_, err = db.DB.Exec(`UPDATE users SET password = ? WHERE id = ?`, hashedPassword, userID)
			if err != nil {
				slog.Error("failed to update user password", "error", err)
				internalServerErrorHandler(w, r, fmt.Errorf("failed to update password"))
				return
			}

			http.Redirect(w, r, "/settings", http.StatusFound)
			return
		}

		if r.FormValue("feed_id") != "" {
			feedID := r.FormValue("feed_id")

			_, err := db.DB.Exec(`DELETE FROM user_feeds WHERE id = ? AND user_id = ?`, feedID, userID)
			if err != nil {
				slog.Error("failed to delete user rss feed", "error", err)

				internalServerErrorHandler(w, r, fmt.Errorf("failed to delete user rss feed"))

				return
			}

			http.Redirect(w, r, "/settings", http.StatusFound)

			return
		}

		if r.FormValue("refresh_interval") != "" {
			intervalStr := r.FormValue("refresh_interval")

			interval, err := strconv.Atoi(intervalStr)
			if err != nil || interval < 0 {
				internalServerErrorHandler(w, r, fmt.Errorf("bad refresh interval"))

				return
			}

			err = db.SetUserRefreshInterval(userID, interval)
			if err != nil {
				slog.Error("failed to set user refresh interval", "error", err)
				internalServerErrorHandler(w, r, fmt.Errorf("failed to update refresh interval"))

				return
			}

			http.Redirect(w, r, "/settings", http.StatusFound)

			return
		}

		feedURL := r.FormValue("feed_url")

		feeds, err := db.GetUserFeeds(userID)
		if err != nil {
			internalServerErrorHandler(w, r, err)

			return
		}

		for _, feed := range feeds {
			if feed.FeedURL == feedURL {
				slog.Info("User feed already exists", "userID", userID, "feed url", feedURL)

				http.Redirect(w, r, "/settings", http.StatusFound)

				return
			}
		}

		feedTitle := feeder.ExtractSourceFromURL(feedURL)

		_, err = db.DB.Exec(`INSERT INTO user_feeds (user_id, feed_url, title) VALUES (?, ?, ?)`, userID, feedURL, feedTitle)
		if err != nil {
			slog.Error("failed to add user feeed rss", "error", err)

			internalServerErrorHandler(w, r, fmt.Errorf("failed to add user rss feed"))

			return
		}

		feedUrls, err := db.GetUserFeedUrls(userID)
		if err != nil {
			internalServerErrorHandler(w, r, err)

			return
		}

		feeder.FetchAndSaveFeeds(feedUrls)

		http.Redirect(w, r, "/settings", http.StatusFound)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

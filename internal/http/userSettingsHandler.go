package http

import (
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const userSettingsTemplate = "templates/userSettings"

func userSettingsRender(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user id from ctx: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	userFeeds, err := db.GetUserFeeds(userInfo.ID)
	if err != nil {
		log.Error("failed to get user feeds: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	refreshInterval, err := db.GetUserRefreshInterval(userInfo.ID)
	if err != nil {
		log.Error("failed to get user refresh interval: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Render(userSettingsTemplate, fiber.Map{
		"UserFeeds":       userFeeds,
		"User":            userInfo,
		"Title":           "RapidFeed - Settings",
		"RefreshInterval": refreshInterval,
	})
}

func changePasswordHandler(c *fiber.Ctx) error {
	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")

	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user info from session: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	hash, err := db.GetUserHash(userInfo.Username)
	if err != nil {
		log.Error("failed to get user hash: ", err)
	}

	err = auth.CheckPassword(hash, currentPassword)
	if err != nil {
		log.Error("wrong current password")
		//TODO: add alert on settings page like in login page, to clearly show where user was wrong
		return c.Redirect("/settings", http.StatusConflict)
	}

	err = db.ChangeUserPassword(userInfo.ID, newPassword)
	if err != nil {
		log.Error("failed to change user password: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/settings", http.StatusFound)
}

//func userSettingsHandler(w http.ResponseWriter, r *http.Request) {
//	switch r.Method {
//		if r.FormValue("feed_id") != "" {
//			feedID := r.FormValue("feed_id")
//
//			_, err := db.DB.Exec(`DELETE FROM user_feeds WHERE id = ? AND user_id = ?`, feedID, userID)
//			if err != nil {
//				slog.Error("failed to delete user rss feed", "error", err)
//
//				internalServerErrorHandler(w, r, fmt.Errorf("failed to delete user rss feed"))
//
//				return
//			}
//
//			http.Redirect(w, r, "/settings", http.StatusFound)
//
//			return
//		}
//
//		if r.FormValue("refresh_interval") != "" {
//			intervalStr := r.FormValue("refresh_interval")
//
//			interval, err := strconv.Atoi(intervalStr)
//			if err != nil || interval < 0 {
//				internalServerErrorHandler(w, r, fmt.Errorf("bad refresh interval"))
//
//				return
//			}
//
//			err = db.SetUserRefreshInterval(userID, interval)
//			if err != nil {
//				slog.Error("failed to set user refresh interval", "error", err)
//				internalServerErrorHandler(w, r, fmt.Errorf("failed to update refresh interval"))
//
//				return
//			}
//
//			http.Redirect(w, r, "/settings", http.StatusFound)
//
//			return
//		}
//
//		feedURL := r.FormValue("feed_url")
//
//		feeds, err := db.GetUserFeeds(userID)
//		if err != nil {
//			internalServerErrorHandler(w, r, err)
//
//			return
//		}
//
//		for _, feed := range feeds {
//			if feed.FeedURL == feedURL {
//				slog.Info("User feed already exists", "userID", userID, "feed url", feedURL)
//
//				http.Redirect(w, r, "/settings", http.StatusFound)
//
//				return
//			}
//		}
//
//		feedTitle := feeder.ExtractSourceFromURL(feedURL)
//
//		_, err = db.DB.Exec(`INSERT INTO user_feeds (user_id, feed_url, title) VALUES (?, ?, ?)`, userID, feedURL, feedTitle)
//		if err != nil {
//			slog.Error("failed to add user feeed rss", "error", err)
//
//			internalServerErrorHandler(w, r, fmt.Errorf("failed to add user rss feed"))
//
//			return
//		}
//
//		feedUrls, err := db.GetUserFeedUrls(userID)
//		if err != nil {
//			internalServerErrorHandler(w, r, err)
//
//			return
//		}
//
//		feeder.FetchAndSaveFeeds(feedUrls)
//
//		http.Redirect(w, r, "/settings", http.StatusFound)
//	default:
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//	}
//}

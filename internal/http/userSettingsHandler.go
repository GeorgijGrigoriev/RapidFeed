package http

import (
	"log/slog"
	"net/http"

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

		user, err := db.GetUserInfo(userID)
		if err != nil {
			internalServerErrorHandler(w, r, err)

			return
		}

		tmpl := PrepareTemplate("internal/templates/base.html",
			"internal/templates/navbar.html",
			"internal/templates/settings.html")

		data := map[string]interface{}{
			"UserFeeds": userFeeds,
			"User":      user,
			"Title":     "Settings - RapidFeed",
		}

		tmpl.ExecuteTemplate(w, "base", data)
	case http.MethodPost:
		if r.FormValue("feed_id") != "" {
			feedID := r.FormValue("feed_id")
			_, err := db.DB.Exec(`DELETE FROM user_feeds WHERE id = ? AND user_id = ?`, feedID, userID)
			if err != nil {
				http.Error(w, "Failed to delete RSS feed", http.StatusInternalServerError)
				return
			}
		} else {
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
				http.Error(w, "Failed to add RSS feed", http.StatusInternalServerError)
				return
			}

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

package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
	"net/http"
)

var settingsTemplate = prepareHTMLTemplate("settings")
var adminUsersTemplate = prepareHTMLTemplate("admin_users")

type UserSettingsHandler struct{}

func (h *UserSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "rapid-feed")
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var userFeeds []models.UserFeed
		rows, err := db.DB.Query(`SELECT id, feed_url FROM user_feeds WHERE user_id = ?`, userID)
		if err != nil {
			http.Error(w, "Failed to fetch RSS feeds", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var feed models.UserFeed
			err := rows.Scan(&feed.ID, &feed.FeedURL)
			if err != nil {
				http.Error(w, "Failed to fetch RSS feeds", http.StatusInternalServerError)
				return
			}
			userFeeds = append(userFeeds, feed)
		}

		settingsTemplate.Execute(w, map[string]interface{}{
			"UserFeeds": userFeeds,
		})
	case http.MethodPost:
		feedURL := r.FormValue("feed_url")
		_, err := db.DB.Exec(`INSERT INTO user_feeds (user_id, feed_url) VALUES (?, ?)`, userID, feedURL)
		if err != nil {
			http.Error(w, "Failed to add RSS feed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings", http.StatusFound)
	case http.MethodDelete:
		feedID := r.FormValue("feed_id")
		_, err := db.DB.Exec(`DELETE FROM user_feeds WHERE id = ? AND user_id = ?`, feedID, userID)
		if err != nil {
			http.Error(w, "Failed to delete RSS feed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings", http.StatusFound)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type AdminUsersHandler struct{}

func (h *AdminUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "rapid-feed")
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	var role string
	err := db.DB.QueryRow(`SELECT role FROM users WHERE id = ?`, userID).Scan(&role)
	if err != nil || role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var users []models.User
		rows, err := db.DB.Query(`SELECT id, username, role FROM users`)
		if err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var user models.User
			err := rows.Scan(&user.ID, &user.Username, &user.Role)
			if err != nil {
				http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}

		var adminUsers []models.AdminUser
		for _, user := range users {
			var userFeeds []models.UserFeed
			rows, err := db.DB.Query(`SELECT id, feed_url FROM user_feeds WHERE user_id = ?`, user.ID)
			if err != nil {
				http.Error(w, "Failed to fetch RSS feeds", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var feed models.UserFeed
				err := rows.Scan(&feed.ID, &feed.FeedURL)
				if err != nil {
					http.Error(w, "Failed to fetch RSS feeds", http.StatusInternalServerError)
					return
				}
				userFeeds = append(userFeeds, feed)
			}

			adminUsers = append(adminUsers, models.AdminUser{
				User:      user,
				UserFeeds: userFeeds,
			})
		}

		adminUsersTemplate.Execute(w, map[string]interface{}{
			"Users": adminUsers,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

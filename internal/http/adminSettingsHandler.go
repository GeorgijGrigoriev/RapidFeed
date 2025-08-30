package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
)

func adminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := checkSession(r)
	if err != nil {
		slog.Debug("attempt to access admin user settings without login")

		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	user, err := db.GetUserInfoById(userID)
	if err != nil {
		internalServerErrorHandler(w, r, err)

		return
	}

	switch r.Method {
	case http.MethodGet:
		usersWithFeeds, err := db.GetUsersWithFeeds()
		if err != nil {
			internalServerErrorHandler(w, r, err)

			return
		}

		tmpl := PrepareTemplate("internal/templates/base.html", "internal/templates/navbar.html", "internal/templates/admin_users.html")

		data := map[string]interface{}{
			"UsersWithFeeds": usersWithFeeds,
			"User":           user,
			"Title":          "Admin - RapidFeed",
		}

		execErr := tmpl.ExecuteTemplate(w, "base", data)
		if execErr != nil {
			slog.Error("failed to execute template", "error", execErr)
		}
	case http.MethodPost:
		username := r.FormValue("username")
		password := r.FormValue("password")
		role := r.FormValue("role")

		if username != "" && password != "" && role != "" {
			hashedPassword, err := auth.HashPassword(password)
			if err != nil {
				slog.Error("failed to hash password", "error", err)

				internalServerErrorHandler(w, r, err)

				return
			}

			_, err = db.DB.Exec(`INSERT INTO users (username, password, role) VALUES (?, ?, ?)`, username, hashedPassword, role)
			if err != nil {
				slog.Error("failed to create new user", "error", err)

				if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
					strings.Contains(err.Error(), "duplicate key") {
					internalServerErrorHandler(w, r, fmt.Errorf("such user already registered"))

					return
				}
				internalServerErrorHandler(w, r, err)

				return
			}

			http.Redirect(w, r, "/admin/users", http.StatusFound)

			return
		}

		blockUserID := r.FormValue("block_user_id")
		if blockUserID != "" {
			userID, err := strconv.Atoi(blockUserID)
			if err != nil {
				slog.Error("invalid user ID for blocking", "error", err)

				internalServerErrorHandler(w, r, fmt.Errorf("invalid user ID"))

				return
			}

			_, err = db.DB.Exec(`UPDATE users SET role = 'blocked' WHERE id = ?`, userID)
			if err != nil {
				slog.Error("failed to block user", "error", err)

				internalServerErrorHandler(w, r, err)

				return
			}

			http.Redirect(w, r, "/admin/users", http.StatusFound)

			return
		}

		deleteFeedID := r.FormValue("delete_feed_id")
		if deleteFeedID != "" {
			feedID, err := strconv.Atoi(deleteFeedID)
			if err != nil {
				slog.Error("invalid feed ID for deletion", "error", err)

				http.Error(w, "Invalid feed ID", http.StatusBadRequest)

				return
			}

			_, err = db.DB.Exec(`DELETE FROM user_feeds WHERE id = ?`, feedID)
			if err != nil {
				slog.Error("failed to delete user feed", "error", err)

				http.Error(w, "Failed to delete user feed", http.StatusInternalServerError)

				return
			}

			http.Redirect(w, r, "/admin/users", http.StatusFound)

			return
		}

		http.Error(w, "Invalid request", http.StatusBadRequest)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

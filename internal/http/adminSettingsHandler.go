package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"log/slog"
	"net/http"
)

func adminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := checkSession(r)
	if err != nil {
		slog.Debug("attempt to access admin user settings without login")

		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	user, err := db.GetUserInfo(userID)
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
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

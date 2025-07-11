package http

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"log/slog"
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "rapid-feed")
		if err != nil {
			slog.Error("session possibly corrupted, creating new one", "error", err)
		}

		userID, ok := session.Values["user_id"].(int)

		slog.Info("user", userID, "ok", ok)

		if !ok || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AdminMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
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
		next.ServeHTTP(w, r)
	})
}

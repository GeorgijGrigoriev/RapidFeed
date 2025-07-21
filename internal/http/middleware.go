package http

import (
	"fmt"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"log/slog"
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, utils.SecretKey)
		if err != nil {
			slog.Error("session possibly corrupted, creating new one", "error", err)
		}

		userID, ok := session.Values["user_id"].(int)

		if !ok || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AdminMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, utils.SecretKey)
		if err != nil {
			slog.Error("session possibly corrupted, creating new one", "error", err)
		}

		userID, ok := session.Values["user_id"].(int)
		if !ok || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusFound)

			return
		}
		role, err := db.GetUserRole(userID)
		if err != nil {
			internalServerErrorHandler(w, r, nil)

			return
		}

		if role != "admin" {
			forbiddenHandler(w, r)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func checkSession(r *http.Request) (int, error) {
	session, _ := store.Get(r, utils.SecretKey)
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		return 0, fmt.Errorf("user session not found")
	}

	return userID, nil
}

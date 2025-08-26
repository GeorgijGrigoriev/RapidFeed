package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
)

const tokenHeaderKey = "X-Token"

func jsonError(code int, err string) string {
	body := map[string]any{"status": code, "error": err}

	marshaledBody, marshalError := json.Marshal(body)
	if marshalError != nil {
		panic(err)
	}

	return string(marshaledBody)
}

func defaultErrorResponser(w http.ResponseWriter, code int, err string) {
	h := w.Header()

	h.Del("Content-Length")

	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	errBody := jsonError(code, err)

	fmt.Fprintln(w, errBody)
}

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

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(tokenHeaderKey)
		if authHeader == "" {
			defaultErrorResponser(w, http.StatusUnauthorized, "missing token")

			return
		}

		tokenInfo, err := db.GetToken(authHeader)
		if err != nil {
			defaultErrorResponser(w, http.StatusUnauthorized, "no token")

			return
		}

		if time.Now().After(tokenInfo.ExpiresAt) {
			defaultErrorResponser(w, http.StatusUnauthorized, "token expired")

			return
		}

		next.ServeHTTP(w, r)
	})
}

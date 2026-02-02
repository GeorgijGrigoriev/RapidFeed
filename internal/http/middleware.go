package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"log/slog"
	"net/http"
	"strings"
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

type contextKey string

const userIDContextKey contextKey = "user_id"

func MCPAuthMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := checkToken(r)
		if err != nil {
			if errors.Is(err, db.ErrTokenNotFound) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			slog.Error("mcp token check failed", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		role, err := db.GetUserRole(userID)
		if err != nil {
			slog.Error("failed to get user role for mcp auth", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if role == "blocked" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
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

func checkToken(r *http.Request) (int, error) {
	if userID, ok := userIDFromContext(r); ok {
		return userID, nil
	}

	token := tokenFromRequest(r)
	if token == "" {
		return 0, db.ErrTokenNotFound
	}

	userID, err := db.GetUserIDByToken(token)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func tokenFromRequest(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		token := strings.TrimSpace(authHeader[7:])
		if token != "" {
			return token
		}
	}

	if token := strings.TrimSpace(r.Header.Get("X-MCP-Token")); token != "" {
		return token
	}

	if token := strings.TrimSpace(r.URL.Query().Get("token")); token != "" {
		return token
	}

	return ""
}

func userIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(userIDContextKey).(int)
	return userID, ok
}

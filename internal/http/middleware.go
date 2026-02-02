package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const (
	tokenHeaderKey = "X-Token"
	tokenInfoKey   = "token-info"
)

// checkSessionMiddleware - check is user logged-in session exists and save it to ctx.
func checkSessionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userInfo, err := getSessionInfo(c)
		if err != nil {
			log.Error("failed to get session info: ", err)

			return c.Redirect("/login", http.StatusFound)
		}

		if userInfo.ID == 0 {
			return c.Redirect("/login", http.StatusFound)
		}

		return c.Next()
	}
}

func adminSessionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userInfo, err := getSessionInfo(c)
		if err != nil {
			log.Error("failed to get session info: ", err)
			return c.Redirect("/login", http.StatusFound)
		}

		if userInfo.ID == 0 {
			return c.Redirect("/login", http.StatusFound)
		}

		if userInfo.Role != "admin" {
			return c.Status(http.StatusForbidden).Render(errorTemplate, defaultForbiddenMap())
		}

		return c.Next()
	}
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

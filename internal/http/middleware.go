package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const (
	tokenHeaderKey = "X-Token"
	tokenInfoKey   = "token-info"
)

func defaultErrorResponser(w http.ResponseWriter, code int, err string) {
	h := w.Header()

	h.Del("Content-Length")

	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)

	errBody := jsonError(code, err)

	fmt.Fprintln(w, errBody)
}

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

		if userInfo.Role != models.AdminRole {
			return c.Render(errorTemplate, defaultForbiddenMap())
		}

		return c.Next()
	}
}

func checkSession(r *http.Request) (int, error) {
	session, _ := store.Get(r, utils.SecretKey)
	userID, ok := session.Values["user_id"].(int)

	if !ok || userID == 0 {
		return 0, fmt.Errorf("user session not found")
	}

	return userID, nil
}

func TokenAuthMiddleware(next http.HandlerFunc) http.Handler {
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

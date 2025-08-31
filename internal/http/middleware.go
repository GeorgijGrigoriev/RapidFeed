package http

import (
	"fmt"
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
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

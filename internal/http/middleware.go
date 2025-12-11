package http

import (
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
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

		if userInfo.Role != models.AdminRole {
			return c.Render(errorTemplate, defaultForbiddenMap())
		}

		return c.Next()
	}
}

func tokenMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(tokenHeaderKey)
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		t, err := db.GetToken(authHeader)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Token error",
			})
		}

		t.ExpiresAt

		// Simple token validation - replace with your own logic (e.g., JWT verification)
		if token != secret {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Optionally set user info in context
		// c.Locals("user", "third-party-service")

		return c.Next()
	}
}

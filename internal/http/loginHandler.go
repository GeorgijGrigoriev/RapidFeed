package http

import (
	"log/slog"
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func loginHandler(c *fiber.Ctx) error {
	switch c.Method() {
	case http.MethodPost:
		username := c.FormValue("username")
		password := c.FormValue("password")

		if username == "" || password == "" {
			slog.Error("[http]: auth error - username or password is empty")

			return c.Render("templates/error", defaultAuthRequiredMap())
		}

		userInfo, err := db.GetUserInfoByUsername(username)
		if err != nil {
			slog.Error("failed to get user info by username")

			return c.Render("templates/error", defaultInternalErrorMap(nil))
		}

		if userInfo.ID == 0 {
			return c.Render("templates/login", fiber.Map{
				"RegisterAllowed": utils.RegisterAllowed,
				"Error":           "This combination of username and password was not found.",
			})
		}

		blocked, err := db.CheckUserBlocked(username)
		if err != nil {
			slog.Error("[http]: user blocked", "username", username)

			return c.SendString("blocked")
		}

		if blocked {
			return c.SendString("forbidden")
		}

		storedHash, err := db.GetUserHash(username)
		if err != nil {
			slog.Error("[http]: failed to get hashed pass", "username", username)

			return c.SendString("internal")
		}

		err = auth.CheckPassword(storedHash, password)
		if err != nil {
			slog.Error("[http]: invalid password", "username", username)

			return c.SendString("forbidden")
		}

		sess, err := sessionStore.Get(c)
		if err != nil {
			slog.Error("[http]: failed to get session store")

			return c.SendString("internal")
		}

		sess.Set("user_id", "1")

		if err = sess.Save(); err != nil {
			panic(err)
		}

		return c.Redirect("/", http.StatusFound)
	case http.MethodGet:
		return c.Render("templates/login", fiber.Map{
			"RegisterAllowed": utils.RegisterAllowed,
		})
	}

	return c.Redirect("/login", http.StatusFound)
}

func CheckSessionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := sessionStore.Get(c)
		if err != nil {
			slog.Error("no session storage found")

			return c.Redirect("/login", http.StatusFound)
		}

		userID := sess.Get("user_id")

		if userID == nil {
			return c.Redirect("/login", http.StatusFound)
		}

		c.Set("user_id", userID.(string))

		return c.Next()
	}
}

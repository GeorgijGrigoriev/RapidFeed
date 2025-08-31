package http

import (
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const loginTemplate = "templates/login"

func loginHandler(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		log.Error("username or password is empty")

		return c.Render(errorTemplate, defaultAuthRequiredMap())
	}

	userInfo, err := db.GetUserInfoByUsername(username)
	if err != nil {
		log.Error("failed to get user info by username", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	if userInfo.ID == 0 {
		return c.Render(loginTemplate, fiber.Map{
			"RegisterAllowed": utils.RegisterAllowed,
			"Error":           "This combination of username and password was not found.",
		})
	}

	if userInfo.Role == models.BlockedRole {
		return c.Render(loginTemplate, fiber.Map{
			"RegisterAllowed": utils.RegisterAllowed,
			"Error":           "Sorry, you have been blocked. Please contact the system administrator for more details.",
		})
	}

	storedHash, err := db.GetUserHash(username)
	if err != nil {
		log.Error("failed to get password hash", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	err = auth.CheckPassword(storedHash, password)
	if err != nil {
		return c.Render(loginTemplate, fiber.Map{
			"RegisterAllowed": utils.RegisterAllowed,
			"Error":           "Wrong username or password.",
		})
	}

	err = saveSessionInfo(c, userInfo)
	if err != nil {
		log.Error("failed to save session", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/", http.StatusFound)
}

func loginRender(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user id from ctx but here ok: ", err)

		return c.Render(loginTemplate, fiber.Map{
			"RegisterAllowed": utils.RegisterAllowed,
		})
	}

	if userInfo.ID != 0 {
		return c.Redirect("/", http.StatusFound)
	}

	return c.Render(loginTemplate, fiber.Map{
		"RegisterAllowed": utils.RegisterAllowed,
	})
}

package http

import (
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const registerTemplate = "templates/register"

func registerHandler(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		log.Error("username or password is empty")

		return c.Render(registerTemplate, fiber.Map{
			"Error": "Username or password must not be empty.",
		})
	}

	userInfo, err := db.GetUserInfoByUsername(username)
	if err != nil {
		log.Error("failed to get user info")
	}

	if userInfo == nil || userInfo.ID != 0 {
		return c.Render(registerTemplate, fiber.Map{
			"Error": "Username already exists",
		})
	}

	err = db.RegisterUser(username, password)
	if err != nil {
		log.Error("failed to register user", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/login", http.StatusFound)
}

func registerRender(c *fiber.Ctx) error {
	return c.Render("templates/register", nil)
}

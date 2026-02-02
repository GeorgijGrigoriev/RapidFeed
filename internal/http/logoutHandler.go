package http

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func logoutHandler(c *fiber.Ctx) error {
	sess, err := sessionStore.Get(c)
	if err != nil {
		log.Error("failed to get session store", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	err = sess.Destroy()
	if err != nil {
		log.Error("failed to logout user", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	return c.Redirect("/login", http.StatusFound)
}

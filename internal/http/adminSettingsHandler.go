package http

import (
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const adminSettingsTemplate = "templates/admin_settings"

func adminSettingsRender(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user id from ctx: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(err))
	}

	usersWithFeeds, err := db.GetUsersWithFeeds()
	if err != nil {
		log.Error("failed to get users and feeds")

		return c.Render(errorTemplate, defaultInternalErrorMap(err))
	}

	return c.Render(adminSettingsTemplate, fiber.Map{
		"UsersWithFeeds": usersWithFeeds,
		"User":           userInfo,
		"Title":          "RapidFeed - Admin settings",
	})
}

func addUserHandler(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	role := c.FormValue("role")

	if username == "" || password == "" || role == "" {
		log.Error("username or password or role is empty but shouldn't")

		return c.Redirect("/admin/users", http.StatusConflict)
	}

	err := db.AddUser(username, password, role)
	if err != nil {
		log.Error("failed to create new user: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(err))
	}

	return c.Redirect("/admin/users", http.StatusFound)
}

func blockUserHandler(c *fiber.Ctx) error {
	blockUserId := c.FormValue("block_user_id")
	if blockUserId == "" {
		log.Warn("empty user id for block is passed, nothing to block")

		return c.Redirect("/admin/users", http.StatusFound)
	}

	err := db.BlockUser(blockUserId)
	if err != nil {
		log.Errorf("failed to block user %s: %v", blockUserId, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(err))
	}

	return c.Redirect("/admin/users", http.StatusFound)
}

func unblockUserHandler(c *fiber.Ctx) error {
	unblockUserId := c.FormValue("unblock_user_id")
	if unblockUserId == "" {
		log.Warn("empty user id for unblock is passed, nothing to unblock")

		return c.Redirect("/admin/users", http.StatusFound)
	}

	err := db.UnblockUser(unblockUserId)
	if err != nil {
		log.Errorf("failed to unblock user %s: %v", unblockUserId, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(err))
	}

	return c.Redirect("/admin/users", http.StatusFound)
}

func removeUserFeedHandler(c *fiber.Ctx) error {
	deleteFeedId := c.FormValue("delete_feed_id")
	if deleteFeedId == "" {
		log.Warn("empty user feed id for dele is passed, nothing to delete")

		return c.Redirect("/admin/users", http.StatusFound)
	}

	err := db.DeleteUserFeed(deleteFeedId)
	if err != nil {
		log.Error("failed to delete user feed: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(err))
	}

	return c.Redirect("/admin/users", http.StatusFound)
}

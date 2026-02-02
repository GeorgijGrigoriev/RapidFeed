package http

import (
	"net/http"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func refreshHandler(c *fiber.Ctx) error {
	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user id from ctx: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	userFeeds, err := db.GetUserFeedUrls(userInfo.ID)
	if err != nil {
		log.Error("failed to get user feed urls: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	feeder.FetchAndSaveFeeds(userFeeds)

	return c.Redirect("/", http.StatusFound)
}

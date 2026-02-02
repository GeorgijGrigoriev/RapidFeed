package http

import (
	"database/sql"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func feedsPageHandler(c *fiber.Ctx) error {
	var (
		items      []models.FeedItem
		page       int
		perPage    int
		totalPages int
		totalCount int
	)

	userInfo, err := getSessionInfo(c)
	if err != nil {
		log.Error("failed to get user id from ctx: ", err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	userFeeds, err := db.GetUserFeedUrls(userInfo.ID)
	if err != nil {
		log.Errorf("failed to get %s feeds: %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	if len(userFeeds) > 0 {
		pageStr := c.Query("page")
		perPageStr := c.Query("per_page")

		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		perPage, err = strconv.Atoi(perPageStr)
		if err != nil || perPage < 1 {
			perPage = 100
		}

		offset := (page - 1) * perPage

		totalCount, err = db.GetTotalUserFeedItemsCount(userFeeds)
		if err != nil {
			log.Errorf("failed to count %s total feed items: %v", userInfo.Username, err)

			return c.Render(errorTemplate, defaultInternalErrorMap(nil))
		}

		totalPages = int(math.Ceil(float64(totalCount) / float64(perPage)))

		items, err = db.GetUserFeedItems(userFeeds, perPage, offset)
		if err != nil {
			log.Errorf("failed to get %s feed items: %v", userInfo.Username, err)

			return c.Render(errorTemplate, defaultInternalErrorMap(nil))
		}
	}

	lastUpdate, err := db.GetLastUpdateTS(userInfo.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("failed to get last update ts for %s feeds: %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	nextUpdate, err := db.GetNextUpdateTS(userInfo.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("failed to get next update ts for %s feeds: %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

	luStr := lastUpdate.Format(time.DateTime)
	nuStr := nextUpdate.Format(time.DateTime)

	if lastUpdate.IsZero() {
		luStr = "Not performed yet"
	}

	if nextUpdate.IsZero() {
		nuStr = "Will be performed soon"
	}

	paginatedItems := models.PaginatedFeedItems{
		Items:      items,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		TotalItems: totalCount,
		LastUpdate: luStr,
		NextUpdate: nuStr,
	}

	return c.Render("templates/index", fiber.Map{
		"PaginatedItems": paginatedItems,
		"User":           userInfo,
		"Title":          "RapidFeed",
		"NoFeeds":        len(userFeeds) == 0,
	})
}

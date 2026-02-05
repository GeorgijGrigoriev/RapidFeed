package http

import (
	"database/sql"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
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

	userFeeds, err := db.GetUserFeeds(userInfo.ID)
	if err != nil {
		log.Errorf("failed to get %s feeds: %v", userInfo.Username, err)

		return c.Render(errorTemplate, defaultInternalErrorMap(nil))
	}

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

	selectedTag := strings.TrimSpace(c.Query("tag"))
	selectedSource := strings.TrimSpace(c.Query("source"))

	availableTags := collectTags(userFeeds)
	filteredFeeds := filterFeeds(userFeeds, selectedTag, selectedSource)
	filteredFeedUrls := extractFeedUrls(filteredFeeds)

	if len(filteredFeedUrls) > 0 {
		offset := (page - 1) * perPage

		totalCount, err = db.GetTotalUserFeedItemsCount(filteredFeedUrls)
		if err != nil {
			log.Errorf("failed to count %s total feed items: %v", userInfo.Username, err)

			return c.Render(errorTemplate, defaultInternalErrorMap(nil))
		}

		totalPages = int(math.Ceil(float64(totalCount) / float64(perPage)))

		items, err = db.GetUserFeedItems(filteredFeedUrls, perPage, offset)
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
		"UserFeeds":      userFeeds,
		"Tags":           availableTags,
		"Filters": fiber.Map{
			"Tag":    selectedTag,
			"Source": selectedSource,
		},
		"User":    userInfo,
		"Title":   "RapidFeed",
		"NoFeeds": len(userFeeds) == 0,
	})
}

func extractFeedUrls(feeds []models.UserFeed) []string {
	urls := make([]string, 0, len(feeds))
	for _, feed := range feeds {
		urls = append(urls, feed.FeedURL)
	}

	return urls
}

func filterFeeds(feeds []models.UserFeed, tag, source string) []models.UserFeed {
	filtered := make([]models.UserFeed, 0, len(feeds))
	tag = strings.TrimSpace(tag)
	source = strings.TrimSpace(source)

	for _, feed := range feeds {
		if source != "" && feed.FeedURL != source {
			continue
		}

		if tag != "" && !feedHasTag(feed, tag) {
			continue
		}

		filtered = append(filtered, feed)
	}

	return filtered
}

func collectTags(feeds []models.UserFeed) []string {
	unique := make(map[string]string)
	for _, feed := range feeds {
		for _, tag := range parseTags(feed.Tags) {
			key := strings.ToLower(tag)
			if _, ok := unique[key]; !ok {
				unique[key] = tag
			}
		}
	}

	tags := make([]string, 0, len(unique))
	for _, tag := range unique {
		tags = append(tags, tag)
	}

	sort.Strings(tags)

	return tags
}

func feedHasTag(feed models.UserFeed, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return false
	}

	for _, tag := range parseTags(feed.Tags) {
		if strings.ToLower(tag) == target {
			return true
		}
	}

	return false
}

func parseTags(tags string) []string {
	parts := strings.Split(tags, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			result = append(result, tag)
		}
	}

	return result
}

package http

import (
	"fmt"
	"log"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"github.com/gofiber/fiber/v2"
)

func feedsPageHandler(c *fiber.Ctx) error {
	var (
		items      []feeder.FeedItem
		page       int
		perPage    int
		totalPages int
		totalCount int
	)

	userID := 1

	userFeeds, err := db.GetUserFeedUrls(userID)
	if err != nil {
		return c.SendString("error")
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

		placeholders := strings.Repeat(",?", len(userFeeds))[1:]

		query := fmt.Sprintf("SELECT COUNT(*) FROM feeds WHERE feed_url IN (%s)", placeholders)
		args := make([]interface{}, len(userFeeds))

		for i, u := range userFeeds {
			args[i] = u
		}

		rows, err := db.DB.Query(query, args...)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&totalCount)
			if err != nil {
				log.Fatal(err)
			}
		}

		totalPages = int(math.Ceil(float64(totalCount) / float64(perPage)))

		argsWithPagination := make([]any, 0, len(userFeeds)+2)
		for _, u := range userFeeds {
			argsWithPagination = append(argsWithPagination, u)
		}

		// Добавляем параметры пагинации
		argsWithPagination = append(argsWithPagination, perPage, offset)
		query = fmt.Sprintf(`SELECT title, link, date, source, description FROM feeds
		WHERE feed_url IN (%s) ORDER BY date DESC LIMIT ? OFFSET ?`, placeholders)

		rows, err = db.DB.Query(query, argsWithPagination...)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		for rows.Next() {
			var item feeder.FeedItem

			err := rows.Scan(&item.Title, &item.Link, &item.Date, &item.Source, &item.Description)
			if err != nil {
				slog.Error("failed to scan feed urls", "error", err)

				continue
			}

			item.Date = timeToHumanReadable(item.Date)

			items = append(items, item)
		}
	}

	lastUpdate, err := db.GetLastUpdateTS(userID)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {

		return c.SendString("internal")
	}

	nextUpdate, err := db.GetNextUpdateTS(userID)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return c.SendString("internal")
	}

	paginatedItems := feeder.PaginatedFeedItems{
		Items:      items,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		TotalItems: totalCount,
		LastUpdate: lastUpdate.Format(time.DateTime),
		NextUpdate: nextUpdate.Format(time.DateTime),
	}

	user, err := db.GetUserInfoById(userID)
	if err != nil {

		return c.SendString("forbidden")
	}

	return c.Render("templates/index", fiber.Map{
		"PaginatedItems": paginatedItems,
		"User":           user,
		"Title":          "RapidFeed",
		"NoFeeds":        len(userFeeds) == 0,
	})
}

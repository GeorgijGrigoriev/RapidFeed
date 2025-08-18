package http

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
)

func handler(tmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	var (
		items      []feeder.FeedItem
		page       int
		perPage    int
		totalPages int
		totalCount int
	)

	userID, err := checkSession(r)
	if err != nil {
		forbiddenHandler(w, r)

		return
	}

	userFeeds, err := db.GetUserFeedUrls(userID)
	if err != nil {
		internalServerErrorHandler(w, r, err)

		return
	}

	if len(userFeeds) > 0 {
		pageStr := r.URL.Query().Get("page")
		perPageStr := r.URL.Query().Get("per_page")

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
	if err != nil {
		internalServerErrorHandler(w, r, err)

		return
	}

	nextUpdate, err := db.GetNextUpdateTS(userID)
	if err != nil {
		internalServerErrorHandler(w, r, err)

		return
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

	user, err := db.GetUserInfo(userID)
	if err != nil {
		forbiddenHandler(w, r)

		return
	}

	data := map[string]interface{}{
		"PaginatedItems": paginatedItems,
		"User":           user,
		"Title":          "RapidFeed",
		"NoFeeds":        len(userFeeds) == 0,
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		slog.Error("failed to execute index page template", "error", err)

		internalServerErrorHandler(w, r, err)

		return
	}
}

func timeToHumanReadable(t string) string {
	time, err := time.Parse(time.RFC3339, t)
	if err != nil {
		slog.Error("failed to parse time", "error", err)

		return t
	}

	return time.Format("2006-01-02 15:04:05")
}

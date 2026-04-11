package feeder

import (
	"log"
	"log/slog"
	"time"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"github.com/mmcdole/gofeed"
)

var feedParser = gofeed.NewParser()

func FetchAndSaveFeeds(urls []string) {
	for _, url := range urls {
		slog.Info("[FEEDER]", "fetching feed", url)

		source := ExtractSourceFromURL(url)
		fetchAndSaveFeed(url, source)
	}
}

func fetchAndSaveFeed(url, source string) {
	fp, err := feedParser.ParseURL(url)
	if err != nil {
		log.Println("Error parsing feed:", err)

		return
	}

	for _, item := range fp.Items {
		var exists bool

		err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM feeds WHERE link = $1 AND feed_url = $2)`, item.Link, url).Scan(&exists)
		if err != nil {
			slog.Error("Error checking not new item in feed", "error", err)

			continue
		}

		if !exists {
			date := item.PublishedParsed.Format(time.RFC3339)

			insertQuery := `INSERT INTO feeds (title, link, date, source, description, feed_url) VALUES (?, ?, ?, ?, ?, ?)`
			title := utils.StripHTMLAndNormalizeFeedText(item.Title)
			description := utils.StripHTMLAndNormalizeFeedText(item.Description)
			normalizedSource := utils.StripHTMLAndNormalizeFeedText(source)

			_, err := db.DB.Exec(insertQuery, title, item.Link, date, normalizedSource, description, url)
			if err != nil {
				slog.Error("Error inserting new item in feed:", "error", err)
			}
		}
	}
}

func ExtractSourceFromURL(url string) string {
	host := ""
	if parsedURL, err := feedParser.ParseURL(url); err == nil && parsedURL.Title != "" {
		host = parsedURL.Title
	}

	return host
}

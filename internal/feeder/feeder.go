package feeder

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/mmcdole/gofeed"
	"log"
	"log/slog"
	"regexp"
)

var feedParser = gofeed.NewParser()

func FetchAndSaveFeeds(urls []string) {
	for _, url := range urls {
		slog.Info("[FEEDER]", "fetching feed", url)

		source := extractSourceFromURL(url)
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

		err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM feeds WHERE link = $1)`, item.Link).Scan(&exists)
		if err != nil {
			log.Println("Error checking not new item in feed:", err)

			continue
		}
		if !exists {
			insertQuery := `INSERT OR IGNORE INTO feeds (title, link, date, source, description) VALUES (?, ?, ?, ?, ?)`
			_, err := db.DB.Exec(insertQuery, item.Title, item.Link, item.PublishedParsed.Format("2006-01-02 15:04:05"), source, cleanHTMLTags(item.Description))
			if err != nil {
				slog.Error("Error inserting new item in feed:", err)
			}
		}
	}
}

func extractSourceFromURL(url string) string {
	host := ""
	if parsedURL, err := feedParser.ParseURL(url); err == nil && parsedURL.Title != "" {
		host = parsedURL.Title
	}

	return host
}

func cleanHTMLTags(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	result := re.ReplaceAllString(input, "")

	return result
}

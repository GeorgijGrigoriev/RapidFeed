package main

import (
	"database/sql"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/http"
	"log"
	"log/slog"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	_ "github.com/mattn/go-sqlite3"
)

var feeds = []string{
	"https://www.cnews.ru/inc/rss/news.xml",
	"https://www.opennet.ru/opennews/opennews_all_utf.rss",
}

func init() {
	slog.Info("Initializing RapidFeed")
	var err error
	db.DB, err = sql.Open("sqlite3", "./feeds.db")
	if err != nil {
		log.Fatal(err)
	}
	createTableQuery := `CREATE TABLE IF NOT EXISTS feeds (
        "id" INTEGER PRIMARY KEY AUTOINCREMENT,
        "title" TEXT,
        "link" TEXT,
        "date" TEXT,
        "source" TEXT
    )`
	db.DB.Exec(createTableQuery)

	feeder.FetchAndSaveFeeds(feeds)
}

func main() {
	slog.Info("Starting RapidFeed server")
	http.New(feeds)
}

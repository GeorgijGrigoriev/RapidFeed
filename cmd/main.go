package main

import (
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
	"https://3dnews.ru/breaking/rss/",
}

func init() {
	slog.Info("Initializing RapidFeed")

	slog.Info("Try to open database")

	db.InitDB()

	slog.Info("Database initialized")

	db.InitSchema() // maybe not necessary call it every time?

	// Create default admin user
	var adminExists bool
	err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = 'admin')`).Scan(&adminExists)
	if err != nil {
		log.Fatal(err)
	}
	if !adminExists {
		insertAdminQuery := `INSERT INTO users (username, password, role) VALUES ('admin', 'admin', 'admin')`
		db.DB.Exec(insertAdminQuery)
	}

	feeder.FetchAndSaveFeeds(feeds)
}

func main() {
	slog.Info("Starting RapidFeed server")
	http.New(feeds)
}

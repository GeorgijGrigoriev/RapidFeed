package main

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/http"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"log/slog"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

var feeds = []string{
	"https://www.cnews.ru/inc/rss/news.xml",
	"https://www.opennet.ru/opennews/opennews_all_utf.rss",
	"https://3dnews.ru/breaking/rss/",
}

func init() {
	slog.Info("Initializing RapidFeed", "version", "1.0.0")

	// Load config vars from env with default fallback
	utils.Listen = utils.GetStringEnv("LISTEN", ":8080")
	utils.SecretKey = utils.GetStringEnv("SECRET_KEY", "strong-secretkey")
	utils.RegisterAllowed = utils.GetBoolEnv("REGISTRATION_ALLOWED", true)

	slog.Info("Try to open database")

	db.InitDB()

	slog.Info("Database initialized")

	db.InitSchema() // maybe not necessary call it every time?

	db.CreateDefaultAdmin()

	slog.Info("Database initialized")
}

func main() {
	slog.Info("Starting RapidFeed server")
	http.New()
}

package main

import (
	"github.com/GeorgijGrigoriev/RapidFeed/internal/http"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/mcp"
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
	utils.MCPListen = utils.GetStringEnv("MCP_LISTEN", ":8090")
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
	go func() {
		slog.Info("Starting RapidFeed MCP server", "listen", utils.MCPListen)
		if err := mcp.Start(utils.MCPListen); err != nil {
			slog.Error("MCP server failed", "error", err)
		}
	}()
	http.New()
}

package main

import (
	"log/slog"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/http"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/mcp"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

var (
	Version = "latest"
	Commit  = "ffffff"
)

func init() {
	slog.Info("Initializing RapidFeed", "version", Version, "commit", Commit)

	// Load config vars from env with default fallback
	utils.Listen = utils.GetStringEnv("LISTEN", ":8080")
	utils.MCPListen = utils.GetStringEnv("MCP_LISTEN", ":8090")
	utils.SecretKey = utils.GetStringEnv("SECRET_KEY", "strong-secretkey")
	utils.RegisterAllowed = utils.GetBoolEnv("REGISTRATION_ALLOWED", true)
	utils.DBPath = utils.GetStringEnv("DB_PATH", "./feeds.db")

	slog.Info("Try to open database")

	db.InitDB()

	slog.Info("Connection opened")

	db.InitSchema() // maybe not necessary call it every time?

	db.CreateDefaultAdmin()

	slog.Info("Database initialized")

	slog.Info("Running database migrations")

	db.RunAllMigrations()

	slog.Info("Database migrations done")
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

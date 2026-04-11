package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/http"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/mcp"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	_ "modernc.org/sqlite"
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

	db.InitDB(utils.DBPath)

	slog.Info("Connection opened")

	slog.Info("Running database migrations")

	if err := db.RunMigrations(db.MigrateUp, 0); err != nil {
		slog.Error("failed to run database migrations", "error", err)

		os.Exit(1)
	}

	slog.Info("Database migrations done")

	db.CreateDefaultAdmin()

	slog.Info("Database initialized")
}

func main() {
	migrateNormalize := flag.Bool("migrate-normalize-feeds", false, "Normalize feed text in DB: strip HTML, decode entities, collapse whitespace. Runs and exits.")
	flag.Parse()

	if *migrateNormalize {
		if err := db.MigrateNormalizeFeedText(); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		return
	}

	slog.Info("Starting RapidFeed server")
	go feeder.StartAutoRefresh()
	go func() {
		slog.Info("Starting RapidFeed MCP server", "listen", utils.MCPListen)
		if err := mcp.Start(utils.MCPListen); err != nil {
			slog.Error("MCP server failed", "error", err)
		}
	}()
	http.New()
}

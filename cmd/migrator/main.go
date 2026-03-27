package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"

	_ "modernc.org/sqlite"
)

func main() {
	direction := flag.String("direction", string(db.MigrateUp), "migration direction: up or down")
	steps := flag.Int("steps", 1, "number of down migration steps")

	flag.Parse()

	dbPath := utils.GetStringEnv("DB_PATH", "./feeds.db")

	slog.Info("Initializing DB connection", "dbPath", dbPath)

	db.InitDB(dbPath)
	defer func() {
		if err := db.DB.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	if err := db.RunMigrations(db.MigrationDirection(*direction), *steps); err != nil {
		slog.Error("failed to run migrations", "direction", *direction, "steps", *steps, "error", err)

		os.Exit(1)
	}

	slog.Info("Migrations completed", "direction", *direction, "steps", *steps)
}

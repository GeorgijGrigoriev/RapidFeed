package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"

	_ "modernc.org/sqlite"
)

func main() {
	direction := flag.String("direction", string(db.MigrateUp), "migration direction: up or down")
	steps := flag.Int("steps", 1, "number of steps for down migration (ignored for up)")
	force := flag.Bool("force", false, "skip confirmation prompt for down migrations")

	flag.Parse()

	dbPath := utils.GetStringEnv("DB_PATH", "./feeds.db")

	slog.Info("Initializing DB connection", "dbPath", dbPath)

	db.InitDB(dbPath)
	defer func() {
		if err := db.DB.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	if db.MigrationDirection(*direction) == db.MigrateDown && !*force {
		fmt.Printf("WARNING: Running DOWN migration will permanently delete data (%d step(s)).\nType 'yes' to confirm: ", *steps)

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		if strings.TrimSpace(scanner.Text()) != "yes" {
			slog.Info("Migration cancelled")
			os.Exit(0)
		}
	}

	if err := db.RunMigrations(db.MigrationDirection(*direction), *steps); err != nil {
		slog.Error("failed to run migrations", "direction", *direction, "steps", *steps, "error", err)

		os.Exit(1)
	}

	slog.Info("Migrations completed", "direction", *direction, "steps", *steps)
}

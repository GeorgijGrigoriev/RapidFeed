package main

import (
	"flag"
	"log/slog"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/db"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/feeder"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/http"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	_ "github.com/mattn/go-sqlite3"
)

var (
	Version = "1.0.2"
	Commit  = "ffffff"
)

func init() {
	slog.Info("Initializing RapidFeed", "version", Version, "commit", Commit)

	listenFlag := flag.String("listen", ":8080", "Address to listen on")
	secretKeyFlag := flag.String("secret-key", "strong-secretkey", "Secret key for sessions")
	registerAllowedFlag := flag.Bool("registration-allowed", true, "Allow user registration")

	flag.Parse()

	utils.Listen = utils.GetStringEnv("RAPIDFEED_LISTEN", *listenFlag)
	utils.SecretKey = utils.GetStringEnv("RAPIDFEED_SECRET_KEY", *secretKeyFlag)
	utils.RegisterAllowed = utils.GetBoolEnv("RAPIDFEED_REGISTRATION_ALLOWED", *registerAllowedFlag)

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
		slog.Info("Starting background feeds puller")
		feeder.StartAutoRefresh()
	}()

	http.New()
}

package db

import (
	"log"
	"log/slog"
	"os"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/auth"
	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
)

const defaultPasswordLength = 14

func CreateDefaultAdmin() {
	var adminExists bool

	err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = 'admin')`).Scan(&adminExists)
	if err != nil {
		log.Fatal(err)
	}

	adminPass, err := auth.GeneratePassword(defaultPasswordLength)
	if err != nil {
		slog.Error("failed to generate password for admin", "error", err)

		os.Exit(1)
	}

	encryptedPass, err := auth.HashPassword(utils.GetStringEnv("ADMIN_PASSWORD", adminPass))
	if err != nil {
		slog.Error("failed to hash default admin password", "error", err)

		os.Exit(1)
	}

	if !adminExists {
		insertAdminQuery := `INSERT INTO users (username, password, role) VALUES ('admin', ? , 'admin')`

		_, err = DB.Exec(insertAdminQuery, encryptedPass)
		if err != nil {
			slog.Error("failed to create default admin", "error", err)

			os.Exit(1)
		}

		slog.Info("!!!!!!!!!!")
		slog.Info("created default admin", "password", adminPass)
		slog.Info("this password shown only this time, keep it in safe place")
		slog.Info("!!!!!!!!!!")
	}
}

package db

import (
	"errors"
	"fmt"

	"github.com/GeorgijGrigoriev/RapidFeed/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type MigrationDirection string

const (
	MigrateUp   MigrationDirection = "up"
	MigrateDown MigrationDirection = "down"
)

func RunMigrations(direction MigrationDirection, steps int) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to initialize migrations source: %w", err)
	}

	dbDriver, err := sqlite.WithInstance(DB, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize sqlite migrations driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	switch direction {
	case MigrateUp:
		err = m.Up()
	case MigrateDown:
		if steps <= 0 {
			return errors.New("steps must be greater than 0 for down migrations")
		}

		err = m.Steps(-steps)
	default:
		return fmt.Errorf("unsupported migration direction: %s", direction)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

package db

import (
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
)

// RunMigrations runs all database "up" migrations in sequential order.
// It skips migrations that have already been run.
func RunMigrations(driver database.Driver, migrationsPath string) error {
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations.
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("Database is already up to date")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

// ForceVersion sets the migration version to a specific value and marks it as
// clean. This is useful for recovering from a dirty migration state.
func ForceVersion(driver database.Driver, migrationsPath string, version int) error {
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Force(version); err != nil {
		return fmt.Errorf("failed to force version: %w", err)
	}

	slog.Info("Forced migration version", "version", version)
	return nil
}

// GetMigrationVersion returns the current migration version.
func GetMigrationVersion(driver database.Driver, migrationsPath string) (uint, bool, error) {
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

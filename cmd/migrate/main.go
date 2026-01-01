package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"tiny-bitly/internal/config"
	"tiny-bitly/internal/db"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/joho/godotenv"
)

func main() {
	var command = flag.String("command", "up", "Migration command: up, force, version")
	var forceVersion = flag.Int("version", 0, "Version number to force (required for force command)")
	flag.Parse()

	// Load environment variables from .env file in development only.
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			slog.Warn("No .env file found, using environment variables", "error", err)
		}
	}

	// Initialize config.
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize logging.
	initLogging(cfg)

	// Get environment variables.
	if cfg.PostgresPort == 0 {
		slog.Error("Postgres port is required. Set environment variable.")
		os.Exit(1)
	}
	if cfg.PostgresDB == "" {
		slog.Error("Postgres DB name is required. Set environment variable.")
		os.Exit(1)
	}
	if cfg.PostgresUser == "" {
		slog.Error("Postgres user is required. Set environment variable.")
		os.Exit(1)
	}
	if cfg.PostgresPassword == "" {
		slog.Error("Postgres password is required. Set environment variable.")
		os.Exit(1)
	}

	// Get migrations path from flag or use default
	// Default to internal/db/migrations relative to project root
	var migrationsPath string
	wd, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to get working directory", "error", err)
		os.Exit(1)
	}
	// If we're in cmd/migrate, go up two levels; otherwise assume we're at root
	if filepath.Base(wd) == "migrate" {
		migrationsPath = filepath.Join(wd, "..", "..", "internal", "db", "migrations")
	} else {
		migrationsPath = filepath.Join(wd, "internal", "db", "migrations")
	}

	// Normalize path
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		slog.Error("Failed to resolve migrations path", "error", err, "path", migrationsPath)
		os.Exit(1)
	}

	slog.Info("Running migrations",
		"command", *command,
		"migrationsPath", absPath,
		"postgresPort", cfg.PostgresPort,
		"postgresDB", cfg.PostgresDB,
		"postgresUser", cfg.PostgresUser,
		"postgresPassword", cfg.PostgresPassword,
	)

	// Open connection.
	dbConnection, err := db.OpenConnectionRaw(
		cfg.PostgresPort,
		cfg.PostgresDB,
		cfg.PostgresUser,
		cfg.PostgresPassword,
	)
	if err != nil {
		slog.Error("Failed to open database connection", "error", err)
		os.Exit(1)
	}
	driver, err := postgres.WithInstance(dbConnection, &postgres.Config{})
	if err != nil {
		slog.Error("Failed to init database driver", "error", err)
		os.Exit(1)
	}
	defer driver.Close()

	// Execute the command.
	switch *command {
	case "up":
		if err := db.RunMigrations(driver, absPath); err != nil {
			slog.Error("Migration failed", "error", err)
			os.Exit(1)
		}
	case "version":
		version, dirty, err := db.GetMigrationVersion(driver, absPath)
		if err != nil {
			slog.Error("Failed to get migration version", "error", err)
			os.Exit(1)
		}
		if dirty {
			slog.Warn("Database is in a dirty state", "version", version)
		} else {
			slog.Info("Current migration version", "version", version)
		}
	case "force":
		if *forceVersion < 0 {
			slog.Error("Version must be >= 0. Use -version flag to specify the version number")
			os.Exit(1)
		}
		slog.Warn("Forcing migration version. This should only be used to recover from a dirty state.")
		if err := db.ForceVersion(driver, absPath, *forceVersion); err != nil {
			slog.Error("Failed to force version", "error", err)
			os.Exit(1)
		}
		slog.Info("Successfully forced migration version", "version", *forceVersion)
	default:
		slog.Error("Unknown command", "command", *command)
		os.Exit(1)
	}
}

func initLogging(cfg *config.Config) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}
	handler = slog.NewTextHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}

package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// OpenConnectionRaw opens a raw connection to the server's PostgreSQL database.
// Originally intended for migrations.
func OpenConnectionRaw(port int, dbName string, user string, password string) (*sql.DB, error) {
	url := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", user, password, port, dbName)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// OpenConnectionRaw opens a GORM-wrapped connection to the server's PostgreSQL database.
// Intended for application code.
func OpenConnectionGORM(port int, dbName string, user string, password string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", user, password, dbName, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Configure connection pool for better concurrency handling
	// Get the underlying *sql.DB to configure pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// SetMaxOpenConns: Maximum number of open connections to the database
	// Should be less than PostgreSQL's max_connections setting
	// For high concurrency, use 25-100 connections (PostgreSQL performs best with 100-300 total)
	sqlDB.SetMaxOpenConns(100)

	// SetMaxIdleConns: Maximum number of connections in the idle connection pool
	// Should be less than or equal to SetMaxOpenConns
	// Keeps connections warm for reuse, reducing connection overhead
	sqlDB.SetMaxIdleConns(25)

	// SetConnMaxLifetime: Maximum amount of time a connection may be reused
	// Prevents stale connections and helps with load balancer connection rotation
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// SetConnMaxIdleTime: Maximum amount of time a connection may be idle
	// Closes idle connections to free up resources
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	// Test the connection with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

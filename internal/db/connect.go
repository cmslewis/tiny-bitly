package db

import (
	"database/sql"
	"fmt"

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
	return db, nil
}

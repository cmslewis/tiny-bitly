package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

// OpenConnection opens a connection to the server's PostgreSQL database.
func OpenConnection(port int, dbName string, user string, password string) (database.Driver, error) {
	url := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", user, password, port, dbName)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	return driver, nil
}

package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/db"
	"tiny-bitly/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// URLRecordDatabaseDAO is a database implementation of URLRecordDAO.
type URLRecordDatabaseDAO struct {
	db *gorm.DB
}

// NewURLRecordDatabaseDAO creates a new database DAO instance.
func NewURLRecordDatabaseDAO(dbPort int, dbName string, dbUser string, dbPassword string) (*URLRecordDatabaseDAO, error) {
	dbConnection, err := db.OpenConnectionGORM(dbPort, dbName, dbUser, dbPassword)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err, "dbPort", dbPort, "dbName", dbName, "dbUser", dbUser)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &URLRecordDatabaseDAO{db: dbConnection}, nil
}

func (d *URLRecordDatabaseDAO) Create(ctx context.Context, urlRecord model.URLRecord) (*model.URLRecordEntity, error) {
	// Create new entity
	entity := model.URLRecordEntity{
		Entity:    model.Entity{},
		URLRecord: urlRecord,
	}

	// Use INSERT ... ON CONFLICT DO NOTHING. This detect conflicts without a
	// separate SELECT query. Use the traditional API for Clauses since generics
	// API doesn't support it directly.
	result := d.db.WithContext(ctx).
		Model(&entity).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "short_code"}},
			DoNothing: true,
		}).
		Create(&entity)

	if result.Error != nil {
		slog.Error(
			"Failed to create record in database",
			"error", result.Error,
			"originalUrl", urlRecord.OriginalURL,
			"shortCode", urlRecord.ShortCode,
		)
		return nil, fmt.Errorf("failed to create record in database: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		// An existing record already has the target short_code.
		return nil, apperrors.ErrShortCodeAlreadyInUse
	}

	return &entity, nil
}

func (d *URLRecordDatabaseDAO) GetByShortCode(ctx context.Context, shortCode string) (*model.URLRecordEntity, error) {
	var entity model.URLRecordEntity

	entity, err := gorm.G[model.URLRecordEntity](d.db).
		Where("short_code = ? AND expires_at > ?", shortCode, time.Now()).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Not found is a normal case
			return nil, nil
		}
		// Actual database error
		slog.Error(
			"Failed to query record by short code in database",
			"error", err,
			"shortCode", shortCode,
		)
		return nil, fmt.Errorf("failed to query record by short code in database: %w", err)
	}

	return &entity, nil
}

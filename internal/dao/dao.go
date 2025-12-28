package dao

import (
	"context"
	"tiny-bitly/internal/dao/memory"
	"tiny-bitly/internal/model"
)

// DAO is the main Data-Access Object that contains all entity-specific DAOs.
type DAO struct {
	URLRecordDAO URLRecordDAO
}

// URLRecordDAO defines the interface for URL record data access operations.
type URLRecordDAO interface {
	Create(ctx context.Context, urlRecord model.URLRecord) (*model.URLRecordEntity, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.URLRecordEntity, error)
}

// NewMemoryDAO creates a new DAO instance using the in-memory implementation.
// This is useful for testing and development.
func NewMemoryDAO() *DAO {
	return &DAO{
		URLRecordDAO: memory.NewURLRecordMemoryDAO(),
	}
}

// NewDatabaseDAO creates a new DAO instance using the database implementation.
// TODO: Implement the database DAO.
func NewDatabaseDAO() (*DAO, error) {
	// TODO: Implement the database DAO.
	// return &DAO{
	//     URLRecordDAO: database.NewURLRecordDatabaseDAO(...),
	// }, nil
	return nil, nil
}

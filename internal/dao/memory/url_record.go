package memory

import (
	"context"
	"sync"
	"time"

	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/model"
)

// URLRecordMemoryDAO is an in-memory implementation of ShortenedURLDAO.
type URLRecordMemoryDAO struct {
	mu        sync.RWMutex
	idCounter int64
	entities  map[string]*model.URLRecordEntity // Map from short code to URL Record
}

// NewURLRecordMemoryDAO creates a new in-memory DAO instance.
func NewURLRecordMemoryDAO() *URLRecordMemoryDAO {
	return &URLRecordMemoryDAO{
		idCounter: 1,
		entities:  make(map[string]*model.URLRecordEntity),
	}
}

func (m *URLRecordMemoryDAO) Create(_ctx context.Context, urlRecord model.URLRecord) (*model.URLRecordEntity, error) {
	// Context is not needed for in-memory store, since in-memory store is very fast.

	m.mu.Lock()
	defer m.mu.Unlock()

	// Fail if this short code is already in use by an active record.

	if existingEntity, ok := m.entities[urlRecord.ShortCode]; ok {
		if !existingEntity.IsDeleted() && !existingEntity.IsExpired() {
			// Simulate a DB query that filters by deleted and expired status:
			return nil, apperrors.ErrShortCodeAlreadyInUse
		}
	}

	now := time.Now()
	entity := &model.URLRecordEntity{
		Entity: model.Entity{
			ID:        m.idCounter,
			CreatedAt: now,
			UpdatedAt: now,
		},
		URLRecord: urlRecord,
	}

	m.entities[entity.ShortCode] = entity
	m.idCounter++

	return entity, nil
}

func (m *URLRecordMemoryDAO) GetByShortCode(_ctx context.Context, shortCode string) (*model.URLRecordEntity, error) {
	// Context is not needed for in-memory store, since in-memory store is very fast.

	m.mu.RLock()
	defer m.mu.RUnlock()

	if existingEntity, ok := m.entities[shortCode]; ok {
		if !existingEntity.IsDeleted() && !existingEntity.IsExpired() {
			return existingEntity, nil
		}
	}

	return nil, nil
}

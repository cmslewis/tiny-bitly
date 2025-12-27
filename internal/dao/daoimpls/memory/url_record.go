package memory

import (
	"sync"
	"time"

	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/dao/daotypes"
	"tiny-bitly/internal/model"
)

// URLRecordMemoryDAO is an in-memory implementation of ShortenedURLDAO.
type URLRecordMemoryDAO struct {
	mu        sync.RWMutex
	idCounter int64
	entities  map[int64]*model.URLRecordEntity
}

// NewURLRecordMemoryDAO creates a new in-memory DAO instance.
func NewURLRecordMemoryDAO() *URLRecordMemoryDAO {
	return &URLRecordMemoryDAO{
		idCounter: 1,
		entities:  make(map[int64]*model.URLRecordEntity),
	}
}

func (m *URLRecordMemoryDAO) Create(urlRecord model.URLRecord) (*model.URLRecordEntity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Fail if this short code is already in use by an active record.
	for _, otherEntity := range m.entities {
		if otherEntity.URLRecord.ShortCode == urlRecord.ShortCode &&
			!otherEntity.IsDeleted() &&
			!otherEntity.IsExpired() {
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

	m.entities[m.idCounter] = entity
	m.idCounter++

	return entity, nil
}

func (m *URLRecordMemoryDAO) GetByShortCode(shortCode string) (*model.URLRecordEntity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entity := range m.entities {
		if entity.IsDeleted() || entity.IsExpired() {
			continue
		}
		if entity.ShortCode == shortCode {
			return entity, nil
		}
	}

	return nil, nil
}

var _ daotypes.URLRecordDAO = (*URLRecordMemoryDAO)(nil)

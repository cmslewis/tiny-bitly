package memory

import (
	"sync"
	"time"

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

func (m *URLRecordMemoryDAO) Create(shortenedUrl model.URLRecord) (*model.URLRecordEntity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	entity := &model.URLRecordEntity{
		Entity: model.Entity{
			ID:        m.idCounter,
			CreatedAt: now,
			UpdatedAt: now,
		},
		URLRecord: shortenedUrl,
	}

	m.entities[m.idCounter] = entity
	m.idCounter++

	return entity, nil
}

func (m *URLRecordMemoryDAO) GetByShortCode(shortCode string) (*model.URLRecordEntity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()

	for _, entity := range m.entities {
		if entity.IsDeleted() {
			continue
		}

		isExpired := entity.ExpiresAt.Before(now)
		if isExpired {
			continue
		}

		if entity.ShortCode == shortCode {
			return entity, nil
		}
	}

	return nil, nil
}

var _ daotypes.URLRecordDAO = (*URLRecordMemoryDAO)(nil)

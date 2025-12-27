package daotypes

import "tiny-bitly/internal/model"

// The main Data-Access Object (DAO) that contains all entity-specific DAOs.
type DAO struct {
	URLRecordDAO URLRecordDAO
}

type URLRecordDAO interface {
	Create(urlRecord model.URLRecord) (*model.URLRecordEntity, error)
	GetByShortCode(shortCode string) (*model.URLRecordEntity, error)
}

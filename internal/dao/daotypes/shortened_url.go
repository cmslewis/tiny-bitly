package daotypes

import "tiny-bitly/internal/model"

type URLRecordDAO interface {
	Create(urlRecord model.URLRecord) (*model.URLRecordEntity, error)
	GetByShortCode(shortCode string) (*model.URLRecordEntity, error)
}

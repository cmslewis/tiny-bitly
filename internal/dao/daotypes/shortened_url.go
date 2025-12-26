package daotypes

import "tiny-bitly/internal/model"

type URLRecordDAO interface {
	Create(shortenedUrl model.URLRecord) (*model.URLRecordEntity, error)
	GetByShortURL(originalURL string) (*model.URLRecordEntity, error)
}

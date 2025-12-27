package daotypes

import (
	"context"
	"tiny-bitly/internal/model"
)

// The main Data-Access Object (DAO) that contains all entity-specific DAOs.
type DAO struct {
	URLRecordDAO URLRecordDAO
}

type URLRecordDAO interface {
	Create(ctx context.Context, urlRecord model.URLRecord) (*model.URLRecordEntity, error)
	GetByShortCode(ctx context.Context, shortCode string) (*model.URLRecordEntity, error)
}

package health

import (
	"context"
	"tiny-bitly/internal/dao"
)

// Service handles health check operations.
type Service struct {
	dao dao.DAO
}

// NewService creates a new health service with the provided dependencies.
func NewService(dao dao.DAO) *Service {
	return &Service{
		dao: dao,
	}
}

// CheckHealth verifies that the service is healthy by checking DAO connectivity.
// Returns true if the service is healthy, false otherwise.
func (s *Service) CheckHealth(ctx context.Context) bool {
	// Perform a simple read operation to verify DAO is accessible.
	// Using a non-existent short code to avoid side effects.
	// GetByShortCode returns (nil, nil) for not found, or (entity, nil) for found,
	// or (nil, error) for actual errors. We just need to verify the DAO responds.
	_, err := s.dao.URLRecordDAO.GetByShortCode(ctx, "__health_check__")

	// If there's an error, the DAO is not accessible (connection failure, etc.)
	// If err is nil, the DAO responded successfully (even if record not found)
	return err == nil
}

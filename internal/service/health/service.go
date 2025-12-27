package health

import (
	"context"
	"tiny-bitly/internal/dao"
)

// checkHealth verifies that the service is healthy by checking DAO connectivity.
// Returns true if the service is healthy, false otherwise.
func checkHealth(ctx context.Context, dao dao.DAO) bool {
	// Perform a simple read operation to verify DAO is accessible.
	// Using a non-existent short code to avoid side effects.
	// GetByShortCode returns (nil, nil) for not found, or (entity, nil) for found,
	// or (nil, error) for actual errors. We just need to verify the DAO responds.
	_, err := dao.URLRecordDAO.GetByShortCode(ctx, "__health_check__")

	// If there's an error, the DAO is not accessible (connection failure, etc.)
	// If err is nil, the DAO responded successfully (even if record not found)
	return err == nil
}

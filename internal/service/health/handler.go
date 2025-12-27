package health

import (
	"encoding/json"
	"net/http"
	"tiny-bitly/internal/dao"
)

type HealthResponse struct {
	Status string `json:"status"`
}

// Creates an HTTP handler for GET /health (liveness probe).
// - 200 OK if the service is alive
// Always returns 200 OK as a basic liveness check (service is running).
func NewHandleGetHealth(dao *dao.DAO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{
			Status: "healthy",
		})
	}
}

// Creates an HTTP handler for GET /ready (readiness probe).
// - 200 OK if the service is ready to accept traffic (DAO is accessible)
// - 503 Service Unavailable if the service is not ready
func NewHandleGetReady(dao *dao.DAO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isHealthy := checkHealth(r.Context(), *dao)

		w.Header().Set("Content-Type", "application/json")

		if !isHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(HealthResponse{
				Status: "unhealthy",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{
			Status: "ready",
		})
	}
}

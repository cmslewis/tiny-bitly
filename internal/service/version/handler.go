package version

import (
	"encoding/json"
	"net/http"
	"tiny-bitly/internal/version"
)

// NewGetVersionHandler creates an HTTP handler for GET /version. Returns
// build-time version information including version, commit, and build time.
func NewGetVersionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		versionInfo := version.Info()
		json.NewEncoder(w).Encode(versionInfo)
	}
}

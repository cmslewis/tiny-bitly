package middleware

import (
	"net/http"

	"github.com/didip/tollbooth/v8"
)

// Returns a middleware that rate limits requests per IP address. When the rate
// limit is exceeded, it returns HTTP 429 Too Many Requests.
func RateLimitMiddleware(next http.Handler, requestsPerSecond int, maxBurstSize int) http.Handler {
	// Create a limiter with the specified rate (tollbooth uses float64 for requests per second)
	limiterInstance := tollbooth.NewLimiter(float64(requestsPerSecond), nil)
	limiterInstance.SetBurst(maxBurstSize)

	// Note: Tollbooth v8 automatically checks X-Forwarded-For, X-Real-IP, and
	// RemoteAddr.

	// Return the tollbooth middleware wrapped around our handler.
	// HTTPMiddleware returns a middleware function that wraps the next handler.
	return tollbooth.HTTPMiddleware(limiterInstance)(next)
}

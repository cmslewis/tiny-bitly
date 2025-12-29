package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"tiny-bitly/internal/constants"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPMetrics holds all HTTP-related Prometheus metrics.
type HTTPMetrics struct {
	// RequestTotal counts the total number of HTTP requests by method,
	// endpoint, and status code.
	RequestTotal *prometheus.CounterVec

	// RequestDuration tracks the duration of HTTP requests by method and
	// endpoint.
	RequestDuration *prometheus.HistogramVec
}

// httpMetrics is the global instance of HTTP metrics.
// Metrics are initialized at package load time using promauto for automatic registration.
var httpMetrics = &HTTPMetrics{
	RequestTotal: promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by method, endpoint, and status code",
		},
		[]string{"method", "endpoint", "status_code"},
	),
	RequestDuration: promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests in seconds, labeled by method, endpoint, and status code",
			// Custom buckets optimized for API latency (milliseconds: 1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000)
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method", "endpoint", "status_code"},
	),
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware tracks HTTP request metrics including:
// 1. Total requests by method, endpoint, and status code
// 2. Request duration by method and endpoint
//
// This middleware should be placed after routing but before business logic handlers.
// It tracks all requests including the /metrics endpoint itself.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code.
		rw := newResponseWriter(w)

		// Process the request.
		next.ServeHTTP(rw, r)

		// Calculate duration.
		duration := time.Since(start).Seconds()

		// Extract labels.
		method := r.Method
		endpoint := normalizeEndpoint(r.URL.Path)
		statusCode := strconv.Itoa(rw.statusCode)

		// Record metrics.
		httpMetrics.RequestTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		httpMetrics.RequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)
	})
}

// normalizeEndpoint normalizes the request path to a consistent endpoint pattern.
// This helps group similar requests (e.g., /abc123 and /xyz789 both become /{shortCode}).
// Known endpoints are preserved exactly, while dynamic segments are normalized.
func normalizeEndpoint(path string) string {
	// Handle root path
	if path == "" || path == "/" {
		return "/"
	}

	// Remove leading/trailing slashes and split
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return "/"
	}

	// Check for known endpoint patterns (exact matches)
	firstPart := parts[0]
	if slices.Contains(constants.ReservedPaths, firstPart) {
		return "/" + firstPart
	}

	// Single segment that's not a known endpoint - likely a short code
	if len(parts) == 1 && firstPart != "" {
		return "/{shortCode}"
	}

	// For multi-segment paths, return the normalized pattern
	// This handles edge cases like nested paths
	return "/" + strings.Join(parts, "/")
}

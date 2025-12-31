package middleware

import "net/http"

// SecurityMiddleware returns a middleware that injects security-related headers
// into responses.
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevents MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevents the page from being displayed in a frame
		w.Header().Set("X-Frame-Options", "DENY")

		// Enables XSS filtering in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Enforces HTTPS connections (only effective when served over HTTPS)
		//
		// Note: Strict-Transport-Security is included even when served over
		// HTTP. In production, this service should be behind a TLS-terminating
		// proxy/load balancer, making this header effective.
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

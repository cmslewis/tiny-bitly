package cache

import (
	"sync"
	"time"
)

// CircuitBreaker implements a simple circuit breaker pattern for Redis.
// It tracks failures and opens the circuit after a threshold, preventing
// cascading failures when Redis is down or overloaded.
type CircuitBreaker struct {
	mu sync.RWMutex

	// State: "closed", "open", "half-open"
	state string

	// Failure tracking (used in "closed" state)
	failureCount    int
	lastFailureTime time.Time

	// Success tracking (used in "half-open" state)
	successCount int

	// Configuration
	failureThreshold int           // Open circuit after this many failures
	resetTimeout     time.Duration // Time to wait before attempting half-open
	successThreshold int           // Close circuit after this many successes in half-open
}

// NewCircuitBreaker creates a new circuit breaker with sensible defaults.
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:            "closed",
		failureThreshold: 5,                // Open after 5 consecutive failures
		resetTimeout:     30 * time.Second, // Wait 30s before trying again
		successThreshold: 2,                // Close after 2 successes in half-open
	}
}

// IsOpen returns true if the circuit is open (should not attempt Redis operations).
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == "open" {
		// Check if enough time has passed to try half-open
		if time.Since(cb.lastFailureTime) >= cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			// Double-check after acquiring write lock
			if cb.state == "open" && time.Since(cb.lastFailureTime) >= cb.resetTimeout {
				cb.state = "half-open"
				cb.failureCount = 0
				cb.successCount = 0 // Reset success counter for half-open state
			}
			cb.mu.Unlock()
			cb.mu.RLock()
		}
	}

	return cb.state == "open"
}

// RecordSuccess records a successful operation and may close the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case "half-open":
		// Track successes in half-open state
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			// Close the circuit - Redis is healthy again
			cb.state = "closed"
			cb.failureCount = 0
			cb.successCount = 0
		}
	case "closed":
		// Reset failure count on success in closed state
		cb.failureCount = 0
	}
}

// RecordFailure records a failed operation and may open the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == "half-open" {
		// Failed in half-open - go back to open
		cb.state = "open"
		cb.failureCount = 0
		cb.successCount = 0
	} else if cb.state == "closed" && cb.failureCount >= cb.failureThreshold {
		// Too many failures - open the circuit
		cb.state = "open"
	}
}

// GetState returns the current circuit breaker state (for debugging).
func (cb *CircuitBreaker) GetState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

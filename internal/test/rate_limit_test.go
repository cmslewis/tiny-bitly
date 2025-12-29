package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/middleware"
	"tiny-bitly/internal/service/create"

	"github.com/stretchr/testify/assert"
)

// Demonstrates that rate limiting is working. This test sends a burst of
// requests and verifies that some requests succeed and some fail with 429 Too
// Many Requests.
func TestRateLimiting(t *testing.T) {
	testConfig := config.GetTestConfig(config.Config{
		APIHostname: "http://localhost:8080",
	})

	// Initialize dependencies.
	appDAO := dao.NewMemoryDAO()
	createService := create.NewService(*appDAO, &testConfig)

	// Build router.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /urls", create.NewPostURLHandler(createService))

	// Apply middleware.
	handler := middleware.RequestIDMiddleware(mux)
	handler = middleware.RateLimitMiddleware(handler, testConfig.RateLimitRequestsPerSecond, testConfig.RateLimitBurst)

	// Create test server.
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Load test parameters
	const (
		numRequests    = 50  // Number of requests to send
		concurrency    = 20  // Number of concurrent goroutines
		requestsPerSec = 100 // Target requests per second (to simulate load)
	)

	var (
		successCount int64
		errorCount   int64
		totalLatency int64 // in nanoseconds
		statusCodes  = make(map[int]int64)
		mu           sync.Mutex
	)

	// Channel to control request rate
	rateLimiter := time.NewTicker(time.Second / requestsPerSec)
	defer rateLimiter.Stop()

	// WaitGroup to wait for all requests to complete
	var wg sync.WaitGroup
	startTime := time.Now()

	// Send requests concurrently
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(requestNum int) {
			defer wg.Done()

			// Wait for rate limiter (simulate request rate)
			<-rateLimiter.C

			// Create request
			originalURL := fmt.Sprintf("https://www.example.com/test-%d", requestNum)
			createReq := map[string]string{
				"url": originalURL,
			}
			reqBody, err := json.Marshal(createReq)
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
				return
			}

			// Send request and measure latency
			reqStart := time.Now()
			resp, err := client.Post(server.URL+"/urls", "application/json", bytes.NewBuffer(reqBody))
			latency := time.Since(reqStart).Nanoseconds()
			atomic.AddInt64(&totalLatency, latency)

			if err != nil {
				atomic.AddInt64(&errorCount, 1)
				return
			}
			defer resp.Body.Close()

			// Track status code
			mu.Lock()
			statusCodes[resp.StatusCode]++
			mu.Unlock()

			// Check if request succeeded
			if resp.StatusCode == http.StatusCreated {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&errorCount, 1)
			}
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()
	totalTime := time.Since(startTime)

	// Calculate statistics
	avgLatency := time.Duration(totalLatency / int64(numRequests))
	actualRPS := float64(numRequests) / totalTime.Seconds()

	// Print results
	t.Logf("\n=== Load Test Results ===")
	t.Logf("Total Requests: %d", numRequests)
	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Target RPS: %d", requestsPerSec)
	t.Logf("Actual RPS: %.2f", actualRPS)
	t.Logf("Total Time: %v", totalTime)
	t.Logf("Success Count: %d", successCount)
	t.Logf("Error Count: %d", errorCount)
	t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(numRequests)*100)
	t.Logf("Average Latency: %v", avgLatency)
	t.Logf("Status Code Distribution:")
	mu.Lock()
	for code, count := range statusCodes {
		t.Logf("  %d: %d", code, count)
	}
	mu.Unlock()

	// Assertions to prove no rate limiting
	// If rate limiting were in place, we would expect some 429 (Too Many Requests) responses
	assert.Equal(t, int64(numRequests), successCount+errorCount, "All requests should have completed")
	assert.Greater(t, statusCodes[http.StatusTooManyRequests], int64(0), "Some 429 responses")
	assert.Greater(t, successCount, int64(0), "Some requests should succeed")
}

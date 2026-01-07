package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	BaseURL             string
	ConcurrentUsers     int
	Duration            time.Duration
	ReadRatio           float64       // 0.0 to 1.0, ratio of read requests vs write requests
	UserRequestInterval time.Duration // Time between requests for each user (0 = as fast as possible)
	WarmupDuration      time.Duration
	CooldownDuration    time.Duration
}

type Stats struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	RateLimited        int64
	Timeouts           int64
	ServerErrors       int64
	ClientErrors       int64
	// Separate tracking for read vs write requests
	ReadRequests      int64
	WriteRequests     int64
	ReadSuccessful    int64
	WriteSuccessful   int64
	ReadRateLimited   int64
	WriteRateLimited  int64
	ReadClientErrors  int64
	WriteClientErrors int64
	Latencies         []time.Duration
	mu                sync.Mutex
}

type RequestResult struct {
	Duration   time.Duration
	StatusCode int
	Error      error
	IsWrite    bool // true for write requests, false for read requests
}

func main() {
	// Get API port from environment variable, default to 8080.
	apiPort := 8080
	if portStr := os.Getenv("API_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			apiPort = port
		}
	}
	defaultURL := fmt.Sprintf("http://localhost:%d", apiPort)

	var (
		baseURL             = flag.String("url", defaultURL, "Base URL of the API")
		concurrentUsers     = flag.Int("users", 50, "Number of concurrent users")
		duration            = flag.Duration("duration", 60*time.Second, "Test duration")
		readRatio           = flag.Float64("read-ratio", 0.8, "Ratio of read requests (0.0-1.0). Ignored if -write-only or -read-only is set")
		userRequestInterval = flag.Duration("user-interval", 2*time.Second, "Time between requests for each user (e.g., 2s = 0.5 req/s per user)")
		warmupDuration      = flag.Duration("warmup", 5*time.Second, "Warmup duration")
		cooldownDuration    = flag.Duration("cooldown", 5*time.Second, "Cooldown duration")
		writeOnly           = flag.Bool("write-only", false, "Only perform write requests (creates short URLs)")
		readOnly            = flag.Bool("read-only", false, "Only perform read requests (requires warmup to create codes first)")
		warmupWrites        = flag.Int("warmup-writes", 1000, "Number of short codes to create during warmup for read-only mode")
	)
	flag.Parse()

	// Validate flags
	if *writeOnly && *readOnly {
		log.Fatal("Cannot specify both -write-only and -read-only")
	}

	// Override readRatio based on mode flags
	effectiveReadRatio := *readRatio
	if *writeOnly {
		effectiveReadRatio = 0.0
	} else if *readOnly {
		effectiveReadRatio = 1.0
	}

	if effectiveReadRatio < 0 || effectiveReadRatio > 1 {
		log.Fatal("read-ratio must be between 0.0 and 1.0")
	}

	cfg := Config{
		BaseURL:             *baseURL,
		ConcurrentUsers:     *concurrentUsers,
		Duration:            *duration,
		ReadRatio:           effectiveReadRatio,
		UserRequestInterval: *userRequestInterval,
		WarmupDuration:      *warmupDuration,
		CooldownDuration:    *cooldownDuration,
	}

	requestsPerUserPerSecond := 0.0
	if cfg.UserRequestInterval > 0 {
		requestsPerUserPerSecond = 1.0 / cfg.UserRequestInterval.Seconds()
	}

	fmt.Printf("Load Test Configuration:\n")
	fmt.Printf("  Base URL: %s\n", cfg.BaseURL)
	fmt.Printf("  Concurrent Users: %d\n", cfg.ConcurrentUsers)
	fmt.Printf("  Duration: %v\n", cfg.Duration)
	if *writeOnly {
		fmt.Printf("  Mode: Write-only (100%% writes)\n")
	} else if *readOnly {
		fmt.Printf("  Mode: Read-only (100%% reads)\n")
		fmt.Printf("  Warmup writes: %d short codes\n", *warmupWrites)
	} else {
		fmt.Printf("  Read Ratio: %.1f%%\n", cfg.ReadRatio*100)
	}
	if cfg.UserRequestInterval > 0 {
		fmt.Printf("  Request Interval per User: %v (%.2f req/s per user)\n", cfg.UserRequestInterval, requestsPerUserPerSecond)
		fmt.Printf("  Expected Total RPS: ~%.2f\n", requestsPerUserPerSecond*float64(cfg.ConcurrentUsers))
	} else {
		fmt.Printf("  Request Interval per User: unlimited (as fast as possible)\n")
	}
	fmt.Printf("  Warmup: %v\n", cfg.WarmupDuration)
	fmt.Printf("  Cooldown: %v\n", cfg.CooldownDuration)
	fmt.Println()

	// Check if server is healthy before starting the test.
	fmt.Printf("Checking server health at %s/health...\n", cfg.BaseURL)
	if err := checkHealth(cfg.BaseURL); err != nil {
		log.Fatalf("Health check failed: %v. Please ensure the server is running and accessible.", err)
	}
	fmt.Println("Health check passed.")
	fmt.Println()

	stats := &Stats{
		Latencies: make([]time.Duration, 0, 100000),
	}

	// Warmup phase
	var preWarmedShortCodes []string
	if *readOnly {
		// For read-only mode, do sequential writes to populate the database
		fmt.Printf("Warming up: Creating %d short codes sequentially...\n", *warmupWrites)
		preWarmedShortCodes = warmupSequentialWrites(cfg.BaseURL, *warmupWrites)
		fmt.Printf("Warmup complete. Created %d short codes.\n", len(preWarmedShortCodes))
	} else if cfg.WarmupDuration > 0 {
		fmt.Printf("Warming up for %v...\n", cfg.WarmupDuration)
		runPhase(cfg, stats, cfg.WarmupDuration, true, nil)
		stats.Reset()
		fmt.Println("Warmup complete.")
	}

	// Main test phase
	fmt.Printf("Starting load test for %v...\n", cfg.Duration)
	startTime := time.Now()
	runPhase(cfg, stats, cfg.Duration, false, preWarmedShortCodes)
	endTime := time.Now()
	actualDuration := endTime.Sub(startTime)

	// Cooldown phase
	if cfg.CooldownDuration > 0 {
		fmt.Printf("\nCooling down for %v...\n", cfg.CooldownDuration)
		time.Sleep(cfg.CooldownDuration)
	}

	// Print results
	printResults(stats, actualDuration)
}

func checkHealth(baseURL string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	healthURL := fmt.Sprintf("%s/health", baseURL)
	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("failed to connect to health endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned status %d (expected 200)", resp.StatusCode)
	}

	return nil
}

func (s *Stats) Reset() {
	atomic.StoreInt64(&s.TotalRequests, 0)
	atomic.StoreInt64(&s.SuccessfulRequests, 0)
	atomic.StoreInt64(&s.FailedRequests, 0)
	atomic.StoreInt64(&s.RateLimited, 0)
	atomic.StoreInt64(&s.Timeouts, 0)
	atomic.StoreInt64(&s.ServerErrors, 0)
	atomic.StoreInt64(&s.ClientErrors, 0)
	atomic.StoreInt64(&s.ReadRequests, 0)
	atomic.StoreInt64(&s.WriteRequests, 0)
	atomic.StoreInt64(&s.ReadSuccessful, 0)
	atomic.StoreInt64(&s.WriteSuccessful, 0)
	atomic.StoreInt64(&s.ReadRateLimited, 0)
	atomic.StoreInt64(&s.WriteRateLimited, 0)
	atomic.StoreInt64(&s.ReadClientErrors, 0)
	atomic.StoreInt64(&s.WriteClientErrors, 0)
	s.mu.Lock()
	s.Latencies = s.Latencies[:0]
	s.mu.Unlock()
}

func (s *Stats) Record(result RequestResult) {
	atomic.AddInt64(&s.TotalRequests, 1)

	// Track read vs write
	if result.IsWrite {
		atomic.AddInt64(&s.WriteRequests, 1)
	} else {
		atomic.AddInt64(&s.ReadRequests, 1)
	}

	if result.Error != nil {
		atomic.AddInt64(&s.FailedRequests, 1)
		if isTimeout(result.Error) {
			atomic.AddInt64(&s.Timeouts, 1)
		}
		return
	}

	switch {
	case result.StatusCode == 200 || result.StatusCode == 201 || result.StatusCode == 302:
		atomic.AddInt64(&s.SuccessfulRequests, 1)
		if result.IsWrite {
			atomic.AddInt64(&s.WriteSuccessful, 1)
		} else {
			atomic.AddInt64(&s.ReadSuccessful, 1)
		}
		s.mu.Lock()
		s.Latencies = append(s.Latencies, result.Duration)
		s.mu.Unlock()
	case result.StatusCode == 429:
		atomic.AddInt64(&s.RateLimited, 1)
		atomic.AddInt64(&s.FailedRequests, 1)
		if result.IsWrite {
			atomic.AddInt64(&s.WriteRateLimited, 1)
		} else {
			atomic.AddInt64(&s.ReadRateLimited, 1)
		}
	case result.StatusCode >= 500:
		atomic.AddInt64(&s.ServerErrors, 1)
		atomic.AddInt64(&s.FailedRequests, 1)
	case result.StatusCode >= 400:
		atomic.AddInt64(&s.ClientErrors, 1)
		atomic.AddInt64(&s.FailedRequests, 1)
		if result.IsWrite {
			atomic.AddInt64(&s.WriteClientErrors, 1)
		} else {
			atomic.AddInt64(&s.ReadClientErrors, 1)
		}
	}
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	// Check for various timeout errors
	return fmt.Sprintf("%v", err) == "context deadline exceeded" ||
		fmt.Sprintf("%v", err) == "timeout" ||
		fmt.Sprintf("%v", err) == "i/o timeout"
}

func warmupSequentialWrites(baseURL string, numWrites int) []string {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	shortCodes := make([]string, 0, numWrites)
	fmt.Printf("Creating %d short codes sequentially...\n", numWrites)

	var errors int
	startTime := time.Now()

	for i := 0; i < numWrites; i++ {
		url := fmt.Sprintf("https://example.com/warmup?index=%d&time=%d", i, time.Now().UnixNano())
		result, createdShortCode := makeWriteRequest(client, baseURL, url)

		if result.Error != nil {
			errors++
			if errors <= 3 {
				// Log first few errors as samples
				fmt.Printf("\nError creating code %d: %v\n", i+1, result.Error)
			}
		} else if result.StatusCode == 429 {
			errors++
			if errors <= 3 {
				fmt.Printf("\nRate limited on code %d (429). Consider disabling rate limiting for warmup.\n", i+1)
			}
			// Add a small delay if rate limited
			time.Sleep(100 * time.Millisecond)
		} else if result.StatusCode == 201 && createdShortCode != "" {
			shortCodes = append(shortCodes, createdShortCode)
		} else if result.StatusCode != 201 {
			errors++
			if errors <= 3 {
				// Log first few non-201 statuses as samples
				fmt.Printf("\nUnexpected status %d for code %d\n", result.StatusCode, i+1)
			}
		}

		// Small delay to avoid overwhelming the server and respect rate limits
		// Even in sequential mode, we want to be gentle
		time.Sleep(10 * time.Millisecond)

		// Progress indicator - show every 10 codes and include success count
		if (i+1)%10 == 0 {
			fmt.Printf("\rProgress: %d/%d created, %d errors", len(shortCodes), i+1, errors)
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\nWarmup complete: %d codes created in %v (%.2f codes/sec)\n",
		len(shortCodes), elapsed.Round(time.Second), float64(len(shortCodes))/elapsed.Seconds())

	if errors > 0 {
		fmt.Printf("Warning: %d errors occurred during warmup\n", errors)
	}

	return shortCodes
}

func runPhase(cfg Config, stats *Stats, duration time.Duration, isWarmup bool, preWarmedShortCodes []string) {
	ctx, cancel := context.WithTimeout(context.Background(), duration+10*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.ConcurrentUsers * 2,
			MaxIdleConnsPerHost: cfg.ConcurrentUsers * 2,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Track short codes created during warmup for read testing
	shortCodes := make([]string, 0, 1000)
	if preWarmedShortCodes != nil {
		// Use pre-warmed codes for read-only mode
		shortCodes = append(shortCodes, preWarmedShortCodes...)
	}
	var shortCodesMu sync.Mutex

	// Start workers (each simulates a user)
	for i := 0; i < cfg.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Create a ticker for this user's request interval
			var userTicker *time.Ticker
			if cfg.UserRequestInterval > 0 {
				userTicker = time.NewTicker(cfg.UserRequestInterval)
				defer userTicker.Stop()
			}

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Wait for the user's request interval if configured
					if userTicker != nil {
						<-userTicker.C
					}

					// Decide read vs write based on ratio
					if shouldRead(cfg.ReadRatio) {
						// Read request - use an existing short code
						shortCodesMu.Lock()
						if len(shortCodes) == 0 {
							// No short codes available yet, skip this read
							shortCodesMu.Unlock()
							continue
						}
						// Pick a random existing short code
						idx := rand.Intn(len(shortCodes))
						shortCode := shortCodes[idx]
						shortCodesMu.Unlock()

						result := makeReadRequest(client, cfg.BaseURL, shortCode)
						result.IsWrite = false
						stats.Record(result)
					} else {
						// Write request
						url := fmt.Sprintf("https://example.com/test?worker=%d&time=%d", workerID, time.Now().UnixNano())
						result, createdShortCode := makeWriteRequest(client, cfg.BaseURL, url)
						result.IsWrite = true
						stats.Record(result)
						if createdShortCode != "" {
							// Store short codes from both warmup and main test
							shortCodesMu.Lock()
							if len(shortCodes) < 1000 {
								shortCodes = append(shortCodes, createdShortCode)
							}
							shortCodesMu.Unlock()
						}
					}

					// If no interval is set, yield to allow other goroutines to run
					// but still make requests as fast as possible
					if userTicker == nil {
						// Small yield to prevent completely starving other goroutines
						time.Sleep(1 * time.Millisecond)
					}
				}
			}
		}(i)
	}

	// Wait for duration
	time.Sleep(duration)
	cancel()
	wg.Wait()
}

func shouldRead(readRatio float64) bool {
	// Random decision based on ratio
	return rand.Float64() < readRatio
}

func makeReadRequest(client *http.Client, baseURL, shortCode string) RequestResult {
	start := time.Now()
	url := fmt.Sprintf("%s/%s", baseURL, shortCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return RequestResult{
			Duration:   time.Since(start),
			StatusCode: 0,
			Error:      err,
		}
	}

	// Don't follow redirects - we just want to measure the response time
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return RequestResult{
			Duration:   duration,
			StatusCode: 0,
			Error:      err,
		}
	}
	defer resp.Body.Close()

	return RequestResult{
		Duration:   duration,
		StatusCode: resp.StatusCode,
		Error:      nil,
	}
}

func makeWriteRequest(client *http.Client, baseURL, originalURL string) (RequestResult, string) {
	start := time.Now()
	url := fmt.Sprintf("%s/urls", baseURL)

	body := map[string]string{"url": originalURL}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return RequestResult{
			Duration:   time.Since(start),
			StatusCode: 0,
			Error:      err,
		}, ""
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return RequestResult{
			Duration:   time.Since(start),
			StatusCode: 0,
			Error:      err,
		}, ""
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return RequestResult{
			Duration:   duration,
			StatusCode: 0,
			Error:      err,
		}, ""
	}
	defer resp.Body.Close()

	shortCode := ""
	if resp.StatusCode == 201 {
		var response struct {
			ShortURL string `json:"shortUrl"`
		}
		if bodyBytes, err := io.ReadAll(resp.Body); err == nil {
			if err := json.Unmarshal(bodyBytes, &response); err == nil {
				// Extract short code from short URL
				if len(response.ShortURL) > 0 {
					// Short URL format: http://host:port/shortCode
					parts := bytes.Split([]byte(response.ShortURL), []byte("/"))
					if len(parts) > 0 {
						shortCode = string(parts[len(parts)-1])
					}
				}
			}
		}
	}

	return RequestResult{
		Duration:   duration,
		StatusCode: resp.StatusCode,
		Error:      nil,
	}, shortCode
}

func printResults(stats *Stats, duration time.Duration) {
	total := atomic.LoadInt64(&stats.TotalRequests)
	successful := atomic.LoadInt64(&stats.SuccessfulRequests)
	failed := atomic.LoadInt64(&stats.FailedRequests)
	rateLimited := atomic.LoadInt64(&stats.RateLimited)
	timeouts := atomic.LoadInt64(&stats.Timeouts)
	serverErrors := atomic.LoadInt64(&stats.ServerErrors)
	clientErrors := atomic.LoadInt64(&stats.ClientErrors)

	rps := float64(total) / duration.Seconds()
	successRate := float64(successful) / float64(total) * 100

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Duration:           %v\n", duration.Round(time.Second))
	fmt.Printf("Total Requests:     %d\n", total)
	fmt.Printf("Successful:         %d (%.2f%%)\n", successful, successRate)
	fmt.Printf("Failed:             %d (%.2f%%)\n", failed, float64(failed)/float64(total)*100)
	fmt.Printf("Throughput:         %.2f req/s\n", rps)
	fmt.Println()
	fmt.Printf("Error Breakdown:\n")
	fmt.Printf("  Rate Limited (429): %d\n", rateLimited)
	fmt.Printf("  Timeouts:           %d\n", timeouts)
	fmt.Printf("  Server Errors (5xx): %d\n", serverErrors)
	fmt.Printf("  Client Errors (4xx): %d\n", clientErrors)
	fmt.Println()

	// Read vs Write breakdown
	readRequests := atomic.LoadInt64(&stats.ReadRequests)
	writeRequests := atomic.LoadInt64(&stats.WriteRequests)
	readSuccessful := atomic.LoadInt64(&stats.ReadSuccessful)
	writeSuccessful := atomic.LoadInt64(&stats.WriteSuccessful)
	readRateLimited := atomic.LoadInt64(&stats.ReadRateLimited)
	writeRateLimited := atomic.LoadInt64(&stats.WriteRateLimited)
	readClientErrors := atomic.LoadInt64(&stats.ReadClientErrors)
	writeClientErrors := atomic.LoadInt64(&stats.WriteClientErrors)

	if readRequests > 0 || writeRequests > 0 {
		fmt.Println("Request Type Breakdown:")
		if readRequests > 0 {
			readSuccessRate := float64(readSuccessful) / float64(readRequests) * 100
			fmt.Printf("  Reads:  %d total, %d successful (%.2f%%), %d rate limited, %d client errors\n",
				readRequests, readSuccessful, readSuccessRate, readRateLimited, readClientErrors)
		}
		if writeRequests > 0 {
			writeSuccessRate := float64(writeSuccessful) / float64(writeRequests) * 100
			fmt.Printf("  Writes: %d total, %d successful (%.2f%%), %d rate limited, %d client errors\n",
				writeRequests, writeSuccessful, writeSuccessRate, writeRateLimited, writeClientErrors)
		}
		fmt.Println()
	}

	// Latency statistics
	stats.mu.Lock()
	latencies := make([]time.Duration, len(stats.Latencies))
	copy(latencies, stats.Latencies)
	stats.mu.Unlock()

	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		p50 := percentile(latencies, 0.50)
		p75 := percentile(latencies, 0.75)
		p90 := percentile(latencies, 0.90)
		p95 := percentile(latencies, 0.95)
		p99 := percentile(latencies, 0.99)
		p999 := percentile(latencies, 0.999)
		min := latencies[0]
		max := latencies[len(latencies)-1]
		avg := average(latencies)

		fmt.Println("Latency Statistics (successful requests only):")
		fmt.Printf("  Min:    %v\n", min.Round(time.Microsecond))
		fmt.Printf("  P50:    %v\n", p50.Round(time.Microsecond))
		fmt.Printf("  P75:    %v\n", p75.Round(time.Microsecond))
		fmt.Printf("  P90:    %v\n", p90.Round(time.Microsecond))
		fmt.Printf("  P95:    %v\n", p95.Round(time.Microsecond))
		fmt.Printf("  P99:    %v\n", p99.Round(time.Microsecond))
		fmt.Printf("  P99.9:  %v\n", p999.Round(time.Microsecond))
		fmt.Printf("  Max:    %v\n", max.Round(time.Microsecond))
		fmt.Printf("  Avg:    %v\n", avg.Round(time.Microsecond))
		fmt.Println()
	}

	// Performance indicators
	fmt.Println("Performance Indicators:")
	if successRate < 95 {
		fmt.Printf("  ⚠️  Success rate is below 95%% - system may be overloaded\n")
	}
	if float64(rateLimited)/float64(total) > 0.01 {
		fmt.Printf("  ⚠️  High rate limiting (%.2f%%) - consider increasing rate limits\n", float64(rateLimited)/float64(total)*100)
	}
	if float64(timeouts)/float64(total) > 0.01 {
		fmt.Printf("  ⚠️  High timeout rate (%.2f%%) - database may be overloaded\n", float64(timeouts)/float64(total)*100)
	}
	if float64(serverErrors)/float64(total) > 0.01 {
		fmt.Printf("  ⚠️  High server error rate (%.2f%%) - check application logs\n", float64(serverErrors)/float64(total)*100)
	}
	fmt.Println(strings.Repeat("=", 80))
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	index := int(math.Ceil(float64(len(sorted)) * p))
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func average(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/middleware"
	"tiny-bitly/internal/service/create"
	"tiny-bitly/internal/service/health"
	"tiny-bitly/internal/service/read"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_HealthReadyCreateGetExpire tests the full flow:
// 1. Health and ready checks
// 2. Create a short URL
// 3. Get it (verify redirect)
// 4. Wait for expiration
// 5. Verify it's no longer accessible
func TestIntegration_HealthReadyCreateGetExpire(t *testing.T) {
	// Use a very short TTL for testing (100ms) to avoid long test times
	shortTTL := 200 * time.Millisecond
	testConfig := config.GetTestConfig(config.Config{
		APIHostname:  "http://localhost:8080",
		ShortCodeTTL: shortTTL,
	})

	// Initialize dependencies.
	appDAO := dao.NewMemoryDAO()
	createService := create.NewService(*appDAO, &testConfig)
	readService := read.NewService(*appDAO, &testConfig)
	healthService := health.NewService(*appDAO)

	// Build router.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.NewGetHealthHandler(healthService))
	mux.HandleFunc("GET /ready", health.NewGetReadyHandler(healthService))
	mux.HandleFunc("POST /urls", create.NewPostURLHandler(createService))
	mux.HandleFunc("GET /{shortCode}", read.NewGetURLHandler(readService))

	// Apply middleware.
	handler := middleware.RequestIDMiddleware(mux)

	// Create test server.
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects automatically, we want to check the status code
			return http.ErrUseLastResponse
		},
		Timeout: 5 * time.Second,
	}

	// Step 1: Test health endpoint
	t.Run("Health check", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var healthResp struct {
			Status string `json:"status"`
		}
		err = json.NewDecoder(resp.Body).Decode(&healthResp)
		require.NoError(t, err)
		assert.Equal(t, "healthy", healthResp.Status)
	})

	// Step 2: Test ready endpoint
	t.Run("Ready check", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/ready")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var readyResp struct {
			Status string `json:"status"`
		}
		err = json.NewDecoder(resp.Body).Decode(&readyResp)
		require.NoError(t, err)
		assert.Equal(t, "ready", readyResp.Status)
	})

	// Step 3: Create a short URL
	var shortCode string
	var originalURL = "https://www.example.com/test"
	t.Run("Create short URL", func(t *testing.T) {
		createReq := map[string]string{
			"url": originalURL,
		}
		reqBody, err := json.Marshal(createReq)
		require.NoError(t, err)

		resp, err := client.Post(server.URL+"/urls", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResp struct {
			ShortURL string `json:"shortUrl"`
		}
		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err)
		require.NotEmpty(t, createResp.ShortURL)

		// Extract short code from the short URL
		// Format: http://localhost:8080/{shortCode}
		// Find the last '/' and take everything after it
		slashIndex := strings.LastIndex(createResp.ShortURL, "/")
		require.Greater(t, slashIndex, -1, "Short URL should contain a slash")
		shortCode = createResp.ShortURL[slashIndex+1:]
		require.NotEmpty(t, shortCode)
	})

	// Step 4: Get the short URL (should redirect)
	t.Run("Get short URL before expiration", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/" + shortCode)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should redirect (302)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, originalURL, resp.Header.Get("Location"))
	})

	// Step 5: Wait for expiration
	t.Run("Wait for expiration", func(t *testing.T) {
		// Wait slightly longer than TTL to ensure expiration
		waitTime := shortTTL + 50*time.Millisecond
		time.Sleep(waitTime)
	})

	// Step 6: Verify the short URL is no longer accessible
	t.Run("Get short URL after expiration", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/" + shortCode)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 404 after expiration
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify error message
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Short code does not exist")
	})
}

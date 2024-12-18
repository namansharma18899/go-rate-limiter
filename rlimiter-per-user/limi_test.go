package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockHandler creates a simple handler for testing
func MockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request processed"))
}

// TestPerUserRateLimiter tests the rate limiter middleware
func TestPerUserRateLimiter(t *testing.T) {
	testCases := []struct {
		name               string
		requestsPerUser    int
		maxAllowedRequests int
		requestInterval    time.Duration
	}{
		{
			name:               "Basic Rate Limiting",
			requestsPerUser:    10,
			maxAllowedRequests: 5,
			requestInterval:    time.Millisecond * 10,
		},
		{
			name:               "Strict Rate Limiting",
			requestsPerUser:    15,
			maxAllowedRequests: 10,
			requestInterval:    time.Millisecond * 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Predefined static IPs for testing
			testIPs := []string{
				"192.168.1.100",
				"192.168.1.101",
				"192.168.1.102",
				"192.168.1.103",
				"192.168.1.104",
			}

			var wg sync.WaitGroup
			results := make(chan int, len(testIPs)*tc.requestsPerUser)

			// Create rate limiter middleware
			rLimiterConfig := RateLimiterConfig{
				Rate:  2.0,
				Burst: tc.maxAllowedRequests,
			}
			rateLimitedHandler := perUserRateLimiter(MockHandler, &rLimiterConfig)

			for _, ip := range testIPs {
				wg.Add(1)
				go func(userIP string) {
					defer wg.Done()

					for i := 0; i < tc.requestsPerUser; i++ {
						// Create a mock request with the specific IP
						req := httptest.NewRequest(http.MethodGet, "/", nil)
						req.Header.Set("X-Forwarded-For", userIP)
						req.RemoteAddr = userIP + ":12345"

						// Create a response recorder
						w := httptest.NewRecorder()

						// Execute the rate-limited handler
						rateLimitedHandler.ServeHTTP(w, req)

						// Record the status code
						results <- w.Result().StatusCode

						// Small delay to simulate real-world scenarios
						time.Sleep(tc.requestInterval)
					}
				}(ip)
			}

			// Wait for all goroutines to complete
			wg.Wait()
			close(results)

			// Analyze results
			var allowedRequests, deniedRequests int
			for result := range results {
				if result == http.StatusOK {
					allowedRequests++
				} else if result == http.StatusTooManyRequests {
					deniedRequests++
				}
			}

			// Assertions
			t.Logf("Allowed Requests: %d, Denied Requests: %d", allowedRequests, deniedRequests)

			assert.LessOrEqual(t, allowedRequests, tc.maxAllowedRequests*len(testIPs),
				"Total allowed requests should not exceed max requests per user * number of users")

			assert.Greater(t, deniedRequests, 0,
				"Some requests should be denied by the rate limiter")
		})
	}
}

// Benchmark the rate limiter performance
func BenchmarkRateLimiter(b *testing.B) {
	rLimiterConfig := RateLimiterConfig{
		Rate:  2.0,
		Burst: 10,
	}
	rateLimitedHandler := perUserRateLimiter(MockHandler, &rLimiterConfig)
	testIP := "192.168.1.250"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", testIP)
		req.RemoteAddr = testIP + ":12345"

		w := httptest.NewRecorder()
		rateLimitedHandler.ServeHTTP(w, req)
	}
}

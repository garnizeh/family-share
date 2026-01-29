package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func TestRateLimit_AllowsRequestsUnderLimit(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 10,
		CleanupInterval:   time.Minute,
	})

	clientIP := "192.168.1.1"

	// Should allow first 10 requests
	for i := 0; i < 10; i++ {
		allowed, remaining, _ := limiter.Allow(clientIP)
		if !allowed {
			t.Errorf("Request %d should be allowed, but was blocked", i+1)
		}
		expectedRemaining := 10 - i - 1
		if remaining != expectedRemaining {
			t.Errorf("Request %d: expected %d remaining, got %d", i+1, expectedRemaining, remaining)
		}
	}
}

func TestRateLimit_BlocksRequestsOverLimit(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 5,
		CleanupInterval:   time.Minute,
	})

	clientIP := "192.168.1.2"

	// Exhaust the limit
	for i := 0; i < 5; i++ {
		allowed, _, _ := limiter.Allow(clientIP)
		if !allowed {
			t.Fatalf("Request %d should be allowed", i+1)
		}
	}

	// 6th request should be blocked
	allowed, remaining, _ := limiter.Allow(clientIP)
	if allowed {
		t.Error("Request over limit should be blocked")
	}
	if remaining != 0 {
		t.Errorf("Expected 0 remaining tokens, got %d", remaining)
	}
}

func TestRateLimit_LimitResetsAfterTimeWindow(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 3,
		CleanupInterval:   time.Minute,
	})

	clientIP := "192.168.1.3"

	// Exhaust the limit
	for i := 0; i < 3; i++ {
		limiter.Allow(clientIP)
	}

	// Should be blocked now
	allowed, _, _ := limiter.Allow(clientIP)
	if allowed {
		t.Error("Request should be blocked after exhausting limit")
	}

	// Manually refill by advancing time
	limiter.mu.Lock()
	bucket := limiter.clients[clientIP]
	bucket.mu.Lock()
	bucket.lastRefill = time.Now().Add(-time.Minute - time.Second)
	bucket.mu.Unlock()
	limiter.mu.Unlock()

	// Should be allowed again after time window
	allowed, _, _ = limiter.Allow(clientIP)
	if !allowed {
		t.Error("Request should be allowed after time window reset")
	}
}

func TestRateLimit_DifferentIPsTrackedSeparately(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 2,
		CleanupInterval:   time.Minute,
	})

	ip1 := "192.168.1.4"
	ip2 := "192.168.1.5"

	// Exhaust limit for IP1
	limiter.Allow(ip1)
	limiter.Allow(ip1)

	// IP1 should be blocked
	allowed, _, _ := limiter.Allow(ip1)
	if allowed {
		t.Error("IP1 should be blocked after exhausting limit")
	}

	// IP2 should still be allowed
	allowed, _, _ = limiter.Allow(ip2)
	if !allowed {
		t.Error("IP2 should be allowed independently of IP1")
	}
}

func TestRateLimit_MiddlewareReturns429(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 2,
		CleanupInterval:   time.Minute,
	})

	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.6:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should return 200, got %d", i+1, w.Code)
		}
	}

	// 3rd request should return 429
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.6:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", w.Code)
	}

	// Check rate limit headers
	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header should be set")
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header should be set on 429")
	}
}

func TestRateLimit_SetsRateLimitHeaders(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 10,
		CleanupInterval:   time.Minute,
	})

	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.7:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("Expected X-RateLimit-Limit: 10, got %s", w.Header().Get("X-RateLimit-Limit"))
	}

	remaining := w.Header().Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Error("X-RateLimit-Remaining header should be set")
	}

	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Error("X-RateLimit-Reset header should be set")
	}
}

func TestRateLimit_HandlesXForwardedFor(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 2,
		CleanupInterval:   time.Minute,
		TrustedProxyCIDRs: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
	})

	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make requests with X-Forwarded-For header
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:12345" // Proxy IP
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should return 200, got %d", i+1, w.Code)
		}
	}

	// 3rd request should be blocked (same X-Forwarded-For)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", w.Code)
	}
}

func TestRateLimit_LockoutAfterViolations(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 2,
		CleanupInterval:   time.Minute,
		LockoutDuration:   time.Minute,
		MaxViolations:     3,
	})

	clientIP := "192.168.1.8"

	// Exhaust limit (2 requests)
	limiter.Allow(clientIP)
	limiter.Allow(clientIP)

	// Make 3 violations (requests over limit)
	for i := 0; i < 3; i++ {
		allowed, _, _ := limiter.Allow(clientIP)
		if allowed {
			t.Error("Request should be blocked (over limit)")
		}
	}

	// Should now be locked out
	limiter.mu.RLock()
	bucket := limiter.clients[clientIP]
	limiter.mu.RUnlock()

	bucket.mu.Lock()
	isLockedOut := !bucket.lockedUntil.IsZero() && time.Now().Before(bucket.lockedUntil)
	bucket.mu.Unlock()

	if !isLockedOut {
		t.Error("Client should be locked out after max violations")
	}

	// Even after manual refill, should still be locked
	bucket.mu.Lock()
	bucket.tokens = 10
	bucket.mu.Unlock()

	allowed, _, _ := limiter.Allow(clientIP)
	if allowed {
		t.Error("Client should remain locked out despite having tokens")
	}
}

func TestRateLimit_CleanupRemovesStaleEntries(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 5,
		CleanupInterval:   100 * time.Millisecond,
	})

	// Create some client buckets
	limiter.Allow("192.168.1.9")
	limiter.Allow("192.168.1.10")

	// Verify they exist
	limiter.mu.RLock()
	initialCount := len(limiter.clients)
	limiter.mu.RUnlock()

	if initialCount != 2 {
		t.Errorf("Expected 2 clients, got %d", initialCount)
	}

	// Make them stale by backdating lastRefill
	limiter.mu.Lock()
	for _, bucket := range limiter.clients {
		bucket.mu.Lock()
		bucket.lastRefill = time.Now().Add(-15 * time.Minute)
		bucket.mu.Unlock()
	}
	limiter.mu.Unlock()

	// Run cleanup
	limiter.cleanup()

	// Verify they've been removed
	limiter.mu.RLock()
	finalCount := len(limiter.clients)
	limiter.mu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Expected 0 clients after cleanup, got %d", finalCount)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		expectedIP    string
		trustedCIDRs  []netip.Prefix
	}{
		{
			name:       "Direct connection",
			remoteAddr: "192.168.1.100:54321",
			expectedIP: "192.168.1.100",
		},
		{
			name:          "X-Forwarded-For single IP (trusted proxy)",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			expectedIP:    "203.0.113.1",
			trustedCIDRs:  []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
		},
		{
			name:          "X-Forwarded-For multiple IPs (trusted proxy)",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1, 198.51.100.1, 192.0.2.1",
			expectedIP:    "203.0.113.1",
			trustedCIDRs:  []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
		},
		{
			name:         "X-Real-IP header (trusted proxy)",
			remoteAddr:   "10.0.0.1:12345",
			xRealIP:      "203.0.113.2",
			expectedIP:   "203.0.113.2",
			trustedCIDRs: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
		},
		{
			name:          "X-Forwarded-For takes precedence over X-Real-IP (trusted proxy)",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "203.0.113.2",
			expectedIP:    "203.0.113.1",
			trustedCIDRs:  []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
		},
		{
			name:          "Untrusted proxy ignores forwarded headers",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "203.0.113.2",
			expectedIP:    "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req, tt.trustedCIDRs)
			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

func TestRateLimit_HelperFunction(t *testing.T) {
	middleware := RateLimit(5)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Should work like NewRateLimiter
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.11:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should return 200, got %d", i+1, w.Code)
		}
	}

	// 6th should be blocked
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.11:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", w.Code)
	}
}

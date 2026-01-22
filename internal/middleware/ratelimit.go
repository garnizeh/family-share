package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TemplateRenderer interface for rendering HTML templates
type TemplateRenderer interface {
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
}

// RateLimiter implements a token bucket algorithm for rate limiting
type RateLimiter struct {
	mu              sync.RWMutex
	requestsPerMin  int
	clients         map[string]*clientBucket
	cleanupInterval time.Duration
	lockoutDuration time.Duration // optional lockout after violations
	maxViolations   int           // number of violations before lockout
	templateRenderer TemplateRenderer // optional template renderer for nice error pages
}

// clientBucket tracks tokens and violations for a single client (IP)
type clientBucket struct {
	tokens         int
	lastRefill     time.Time
	violations     int
	lockedUntil    time.Time
	mu             sync.Mutex
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int
	CleanupInterval   time.Duration
	LockoutDuration   time.Duration
	MaxViolations     int
	TemplateRenderer  TemplateRenderer // optional template renderer
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}
	if config.MaxViolations == 0 {
		config.MaxViolations = 10 // default: lockout after 10 violations
	}

	rl := &RateLimiter{
		requestsPerMin:   config.RequestsPerMinute,
		clients:          make(map[string]*clientBucket),
		cleanupInterval:  config.CleanupInterval,
		lockoutDuration:  config.LockoutDuration,
		maxViolations:    config.MaxViolations,
		templateRenderer: config.TemplateRenderer,
	}

	// Start background cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// RateLimit creates middleware with specified requests per minute
func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		LockoutDuration:   5 * time.Minute,
		MaxViolations:     10,
	})
	return limiter.Middleware()
}

// Middleware returns an HTTP middleware function
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			
			allowed, remaining, resetTime := rl.Allow(clientIP)
			
			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.requestsPerMin))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
			
			if !allowed {
				log.Printf("Rate limit exceeded for IP: %s on %s", clientIP, r.URL.Path)
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
				
				// Try to render nice HTML error page if template renderer is available
				if rl.templateRenderer != nil {
					w.WriteHeader(http.StatusTooManyRequests)
					data := struct {
						RetryAfter int
					}{
						RetryAfter: int(time.Until(resetTime).Seconds()),
					}
					if err := rl.templateRenderer.RenderTemplate(w, "rate_limit.html", data); err != nil {
						log.Printf("Failed to render rate limit template: %v", err)
						// Fallback to plain text
						http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
					}
					return
				}
				
				// Fallback to plain text error
				http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// Allow checks if a request from the given client IP is allowed
// Returns: (allowed bool, remaining tokens, reset time)
func (rl *RateLimiter) Allow(clientIP string) (bool, int, time.Time) {
	rl.mu.Lock()
	bucket, exists := rl.clients[clientIP]
	if !exists {
		bucket = &clientBucket{
			tokens:     rl.requestsPerMin,
			lastRefill: time.Now().UTC(),
		}
		rl.clients[clientIP] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Check if client is locked out
	if !bucket.lockedUntil.IsZero() && time.Now().UTC().Before(bucket.lockedUntil) {
		resetTime := bucket.lockedUntil
		return false, 0, resetTime
	}

	// Reset lockout if expired
	if !bucket.lockedUntil.IsZero() && time.Now().UTC().After(bucket.lockedUntil) {
		bucket.lockedUntil = time.Time{}
		bucket.violations = 0
	}

	// Refill tokens based on time elapsed
	now := time.Now().UTC()
	elapsed := now.Sub(bucket.lastRefill)
	
	// Token bucket: refill proportionally to time elapsed
	// Full refill happens every minute
	if elapsed >= time.Minute {
		bucket.tokens = rl.requestsPerMin
		bucket.lastRefill = now
	} else {
		// Partial refill based on elapsed time
		tokensToAdd := int(float64(rl.requestsPerMin) * (elapsed.Seconds() / 60.0))
		bucket.tokens += tokensToAdd
		if bucket.tokens > rl.requestsPerMin {
			bucket.tokens = rl.requestsPerMin
		}
		// Update lastRefill only if we added tokens
		if tokensToAdd > 0 {
			bucket.lastRefill = now
		}
	}

	// Check if tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		nextRefill := bucket.lastRefill.Add(time.Minute)
		return true, bucket.tokens, nextRefill
	}

	// No tokens available - record violation
	bucket.violations++
	
	// Apply lockout if too many violations
	if rl.lockoutDuration > 0 && bucket.violations >= rl.maxViolations {
		bucket.lockedUntil = now.Add(rl.lockoutDuration)
		log.Printf("Client %s locked out until %v after %d violations", 
			clientIP, bucket.lockedUntil, bucket.violations)
		return false, 0, bucket.lockedUntil
	}

	nextRefill := bucket.lastRefill.Add(time.Minute)
	return false, 0, nextRefill
}

// cleanupLoop periodically removes stale client buckets to prevent memory leaks
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes client buckets that haven't been used recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now().UTC()
	staleThreshold := 10 * time.Minute

	for ip, bucket := range rl.clients {
		bucket.mu.Lock()
		lastActivity := bucket.lastRefill
		isLocked := !bucket.lockedUntil.IsZero() && now.Before(bucket.lockedUntil)
		bucket.mu.Unlock()

		// Remove if inactive for too long and not locked out
		if !isLocked && now.Sub(lastActivity) > staleThreshold {
			delete(rl.clients, ip)
		}
	}
}

// getClientIP extracts the client IP from the request
// Handles X-Forwarded-For header for reverse proxy scenarios
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by reverse proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, use the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

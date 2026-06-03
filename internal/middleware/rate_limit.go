package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/C9b3rD3vi1/Angazia/internal/pkg/utils"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	GeneralLimit   int
	AuthLimit      int
	SearchLimit    int
	WebSocketLimit int
	AdminLimit     int
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GeneralLimit:   60,
		AuthLimit:      10,
		SearchLimit:    30,
		WebSocketLimit: 5,
		AdminLimit:     120,
	}
}

// RateLimit returns a rate limiter middleware based on path
func RateLimit(config RateLimitConfig) fiber.Handler {
	// General rate limiter
	generalLimiter := limiter.New(limiter.Config{
		Max:        config.GeneralLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "rl:general:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.Error(c, fiber.StatusTooManyRequests, "Too many requests. Please try again later.")
		},
	})

	// Auth rate limiter (stricter)
	authLimiter := limiter.New(limiter.Config{
		Max:        config.AuthLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit by IP + email for login attempts
			email := c.FormValue("email", "")
			return "rl:auth:" + c.IP() + ":" + email
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.Error(c, fiber.StatusTooManyRequests, "Too many authentication attempts. Please try again after 1 minute.")
		},
	})

	// Search rate limiter
	searchLimiter := limiter.New(limiter.Config{
		Max:        config.SearchLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "rl:search:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.Error(c, fiber.StatusTooManyRequests, "Too many search requests. Please slow down.")
		},
	})

	// Admin rate limiter (higher limit)
	adminLimiter := limiter.New(limiter.Config{
		Max:        config.AdminLimit,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("user_id")
			if userID != nil {
				return "rl:admin:" + userID.(string)
			}
			return "rl:admin:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.Error(c, fiber.StatusTooManyRequests, "Admin rate limit exceeded. Please try again later.")
		},
	})

	// Return middleware that applies appropriate limiter based on path
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Auth endpoints
		if contains(path, "/api/v1/auth/login", "/api/v1/auth/register",
			"/api/v1/auth/forgot-password", "/api/v1/auth/reset-password") {
			return authLimiter(c)
		}

		// Search endpoints
		if contains(path, "/api/v1/search/", "/api/v1/jobs/search") {
			return searchLimiter(c)
		}

		// Admin endpoints
		if contains(path, "/api/v1/admin/") {
			return adminLimiter(c)
		}

		// WebSocket upgrade
		if path == "/ws" {
			return websocketRateLimit(config.WebSocketLimit)(c)
		}

		// Default general limiter
		return generalLimiter(c)
	}
}

// IPRateLimiter limits requests by IP address with custom configuration
func IPRateLimiter(max int, expiration time.Duration, message string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "rl:ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.Error(c, fiber.StatusTooManyRequests, message)
		},
	})
}

// UserRateLimiter limits requests by user ID (authenticated users)
func UserRateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("user_id")
			if userID != nil {
				return "rl:user:" + userID.(string)
			}
			return "rl:user:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.Error(c, fiber.StatusTooManyRequests, "Too many requests. Please try again later.")
		},
	})
}

// websocketRateLimit limits WebSocket connections per IP
func websocketRateLimit(max int) fiber.Handler {
	var mu sync.Mutex
	connections := make(map[string]int)

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		mu.Lock()
		defer mu.Unlock()

		if connections[ip] >= max {
			return utils.Error(c, fiber.StatusTooManyRequests, "Too many WebSocket connections from this IP")
		}

		connections[ip]++

		// Clean up on disconnect
		c.Context().SetUserValue("cleanup", func() {
			mu.Lock()
			defer mu.Unlock()
			connections[ip]--
			if connections[ip] <= 0 {
				delete(connections, ip)
			}
		})

		return c.Next()
	}
}

// SlidingWindowRateLimiter implements sliding window rate limiting
type SlidingWindowRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewSlidingWindowRateLimiter(limit int, window time.Duration) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (r *SlidingWindowRateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// Clean up old requests
	requests := r.requests[key]
	valid := make([]time.Time, 0)
	for _, t := range requests {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= r.limit {
		return false
	}

	r.requests[key] = append(valid, now)
	return true
}

// Helper function
func contains(path string, patterns ...string) bool {
	for _, p := range patterns {
		if p == path {
			return true
		}
	}
	return false
}

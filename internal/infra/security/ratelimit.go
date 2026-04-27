package security

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// tokenBucket implements a simple per-key token-bucket rate limiter.
type tokenBucket struct {
	tokens   float64
	last     time.Time
	capacity float64
	refill   float64 // tokens per second
}

// RateLimiterConfig configures the rate limiter middleware.
type RateLimiterConfig struct {
	// RequestsPerSecond is the steady-state rate (refill speed).
	RequestsPerSecond float64
	// Burst is the bucket capacity (max burst).
	Burst float64
	// KeyFunc derives the bucket key from the request. Defaults to RealIP.
	KeyFunc func(c echo.Context) string
}

// RateLimiter returns an Echo middleware enforcing a per-key token bucket.
// It is intentionally lightweight (single process, in-memory). For multi-replica
// deployments use a shared store (Redis) instead.
func RateLimiter(cfg RateLimiterConfig) echo.MiddlewareFunc {
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = 10
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 20
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c echo.Context) string { return c.RealIP() }
	}

	var mu sync.Mutex
	buckets := make(map[string]*tokenBucket)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := cfg.KeyFunc(c)
			now := time.Now()

			mu.Lock()
			b, ok := buckets[key]
			if !ok {
				b = &tokenBucket{tokens: cfg.Burst, last: now, capacity: cfg.Burst, refill: cfg.RequestsPerSecond}
				buckets[key] = b
			}
			elapsed := now.Sub(b.last).Seconds()
			b.tokens = min(b.capacity, b.tokens+elapsed*b.refill)
			b.last = now
			allow := b.tokens >= 1
			if allow {
				b.tokens--
			}
			mu.Unlock()

			if !allow {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

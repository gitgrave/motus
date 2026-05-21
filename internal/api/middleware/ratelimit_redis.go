package middleware

import (
	"net"
	"net/http"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// NewRedisLoginRateLimit returns middleware that enforces the given rate limit
// using Redis as the shared store, making it effective across multiple pods.
// IP extraction follows the same precedence as the in-process limiter:
// X-Forwarded-For → X-Real-Ip → RemoteAddr.
func NewRedisLoginRateLimit(client *redis.Client, cfg RateLimitConfig) func(http.Handler) http.Handler {
	limiter := redis_rate.NewLimiter(client)
	limit := redis_rate.Limit{
		Rate:   int(cfg.Max),
		Period: cfg.Period,
		Burst:  int(cfg.Max),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractClientIP(r)
			key := "motus:ratelimit:login:" + ip

			res, err := limiter.Allow(r.Context(), key, limit)
			if err != nil || res.Allowed == 0 {
				rateLimitResponse(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP returns the client IP from X-Forwarded-For, X-Real-Ip, or
// RemoteAddr, in that order of precedence.
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple IP-based rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // max requests
	window   time.Duration // time window
	done     chan struct{} // channel to signal goroutine stop
}

type visitor struct {
	count     int
	expiresAt time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		done:     make(chan struct{}),
	}
	// Cleanup expired entries periodically
	go rl.cleanup()
	return rl
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists || now.After(v.expiresAt) {
		rl.visitors[ip] = &visitor{
			count:     1,
			expiresAt: now.Add(rl.window),
		}
		return true
	}

	if v.count >= rl.rate {
		return false
	}

	v.count++
	return true
}

// cleanup removes expired entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.done:
			// Stop signal received
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.After(v.expiresAt) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// StopWithContext stops the cleanup goroutine using context
func (rl *RateLimiter) StopWithContext(ctx context.Context) {
	select {
	case <-rl.done:
		// Already closed
	default:
		close(rl.done)
	}
}

// RateLimitMiddleware returns a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

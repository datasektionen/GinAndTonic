package routes

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// userLimiter holds the rate limiter for a user and the last time they were seen
type userLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterMiddleware handles the rate limiting per user
type RateLimiterMiddleware struct {
	visitors map[string]*userLimiter
	mtx      sync.Mutex
	r        rate.Limit
	b        int
}

// NewRateLimiterMiddleware initializes and returns a new instance of RateLimiterMiddleware
func NewRateLimiterMiddleware(r rate.Limit, b int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		visitors: make(map[string]*userLimiter),
		r:        r,
		b:        b,
	}
}

// addVisitor adds a new user or updates the last seen time for an existing user
func (m *RateLimiterMiddleware) addVisitor(userID string) *rate.Limiter {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	v, exists := m.visitors[userID]
	if !exists || time.Now().Sub(v.lastSeen) > 1*time.Minute {
		limiter := rate.NewLimiter(m.r, m.b)
		m.visitors[userID] = &userLimiter{limiter, time.Now()}
		return limiter
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// MiddlewareFunc returns the Gin middleware function
func (m *RateLimiterMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("user_id").(string)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
			return
		}

		limiter := m.addVisitor(userID)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}

func (m *RateLimiterMiddleware) MiddlewareFuncURLParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("ugkthid")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
			return
		}

		limiter := m.addVisitor(userID)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}

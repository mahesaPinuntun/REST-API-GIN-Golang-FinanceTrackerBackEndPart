package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		// 5 requests/sec, burst of 10
		limiter = rate.NewLimiter(5, 10)
		visitors[ip] = limiter
	}

	return limiter
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {

		ip := c.ClientIP()

		limiter := getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
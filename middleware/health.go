package middleware

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

var isHealthy atomic.Bool

func init() {
	isHealthy.Store(true)
}

// SetHealthStatus sets the health status of the service
func SetHealthStatus(healthy bool) {
	isHealthy.Store(healthy)
}

// Health returns a middleware that handles health check requests
func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			if isHealthy.Load() {
				c.JSON(http.StatusOK, gin.H{
					"status": "healthy",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "unhealthy",
				})
			}
			c.Abort()
			return
		}
		c.Next()
	}
}
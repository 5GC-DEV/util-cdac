package middleware

import (
	"net/http"
	"sync"

	"github.com/5GC-DEV/util-cdac/logger"
	"github.com/gin-gonic/gin"
)

var requestCache sync.Map

func IdempotencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		requestInfo := c.GetHeader("3gpp-Sbi-Request-Info")

		if requestInfo != "" {
			logger.UtilLog.Infof(
				"Received idempotency key=%s method=%s path=%s",
				requestInfo,
				c.Request.Method,
				c.Request.URL.Path,
			)
		}

		if requestInfo == "" {
			c.Next()
			return
		}

		if _, exists := requestCache.Load(requestInfo); exists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "duplicate request detected",
			})
			c.Abort()
			return
		}

		requestCache.Store(requestInfo, true)

		c.Next()
	}
}

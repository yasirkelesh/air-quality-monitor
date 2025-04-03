// middleware/logger.go
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Basit bir loglama middleware'i
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// İstek başlamadan önce zaman kaydı
		startTime := time.Now()

		// İsteği işle
		c.Next()

		// İstek tamamlandıktan sonra loglama
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		log.Printf(
			"[API-GATEWAY] %s | %3d | %13v | %15s | %s",
			c.Request.Method,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			c.Request.URL.Path,
		)
	}
}

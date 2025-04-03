// api/routes.go (güncelleme)
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yasirkelesh/data-collector/service"
)

// SetupRoutes API rotalarını yapılandırır
func SetupRoutes(router *gin.Engine, pollutionService *service.PollutionService) {
	// Handler oluştur
	handler := NewPollutionHandler(pollutionService)

	// API versiyonu grubu
	v1 := router.Group("/api/v1")
	{
		// Veri ekleme endpoint'i
		v1.POST("/pollution", handler.AddPollutionData)
		// Veri listeleme endpoint'i
		v1.GET("/pollution", handler.GetPollutionData)

		// Sağlık kontrolü endpoint'i
		v1.GET("/health", handler.GetHealth)
	}

	// Ping endpoint'i
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}

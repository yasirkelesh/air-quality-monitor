// main.go
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yasirkelesh/api-gateway/config"
	"github.com/yasirkelesh/api-gateway/middleware"
	"github.com/yasirkelesh/api-gateway/proxy"
)

func main() {
	// Konfigürasyon yükle
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Gin modunu ayarla
	gin.SetMode(cfg.Server.Mode)

	// Router oluştur
	router := gin.Default()

	// Middleware'leri ekle
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	// Sağlık kontrolü endpoint'i
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "up",
			"service": "api-gateway",
		})
	})

	// Kimlik doğrulama aktifse, auth middleware'ini ekle
	if cfg.Auth.Enabled {
		log.Println("Authentication is enabled")

		// Auth endpoint'i
		router.POST("/auth/login", func(c *gin.Context) {
			// Basit bir login işlemi örneği
			var credentials struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.ShouldBindJSON(&credentials); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// TODO: Gerçek kimlik doğrulama burada yapılmalı

			c.JSON(http.StatusOK, gin.H{
				"token": "sample-jwt-token",
			})
		})

		// Korunan rotalar için auth middleware'i ekle
		// authGroup := router.Group("/api")
		// authGroup.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
		// {
		//     // Korunan rotaları burada tanımla
		// }
	}

	// Veri toplama servisi için proxy
	dataCollectorProxy := proxy.NewServiceProxy(cfg.Services.DataCollector, "data-collector")
	router.Any("/api/data-collector/*proxyPath", dataCollectorProxy.ReverseProxy())

	// Veri işleme servisi için proxy
	dataProcessingProxy := proxy.NewServiceProxy(cfg.Services.DataProcessing, "data-processing")
	router.Any("/api/data-processing/*proxyPath", dataProcessingProxy.ReverseProxy())

	// Anomali tespit servisi için proxy
	anomalyDetectionProxy := proxy.NewServiceProxy(cfg.Services.AnomalyDetection, "anomaly-detection")
	router.Any("/api/anomaly-detection/*proxyPath", anomalyDetectionProxy.ReverseProxy())

	// Bildirim servisi için proxy
	notificationProxy := proxy.NewServiceProxy(cfg.Services.Notification, "notification")
	router.Any("/api/notification/*proxyPath", notificationProxy.ReverseProxy())

	// API gateway'i başlat
	log.Printf("API Gateway starting on port %s in %s mode", cfg.Server.Port, cfg.Server.Mode)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}

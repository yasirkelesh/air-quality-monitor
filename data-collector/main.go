// main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/yasirkelesh/data-collector/api"
	"github.com/yasirkelesh/data-collector/config"
	"github.com/yasirkelesh/data-collector/repository"
	"github.com/yasirkelesh/data-collector/service"
)

func main() {
	// Konfigürasyon yükle
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// MongoDB repository oluştur
	repo, err := repository.NewMongoRepository(
		cfg.MongoDB.URI,
		cfg.MongoDB.Database,
		cfg.MongoDB.Collection,
	)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer repo.Close()

	// Servis katmanını oluştur
	pollutionService := service.NewPollutionService(repo)

	// MQTT istemcisini başlat
	if cfg.MQTT.BrokerURL != "" {
		if err := pollutionService.InitMQTT(
			cfg.MQTT.BrokerURL, 
			cfg.MQTT.ClientID, 
			cfg.MQTT.Topic,
		); err != nil {
			log.Printf("MQTT bağlantısı başlatılamadı: %v", err)
		} else {
			log.Printf("MQTT istemcisi başlatıldı, broker: %s, topic: %s", 
				cfg.MQTT.BrokerURL, cfg.MQTT.Topic)
			defer pollutionService.CloseMQTT()
		}
	}

	// Gin modunu ayarla
	gin.SetMode(cfg.Server.Mode)
	
	// Gin router oluştur
	router := gin.Default()
	
	// CORS middleware ekle
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// API rotalarını ayarla
	api.SetupRoutes(router, pollutionService)
	
	// Graceful shutdown için sinyal yakalama
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-quit
		log.Println("Servis kapatılıyor...")
		pollutionService.CloseMQTT()
		os.Exit(0)
	}()
	
	// Sunucuyu başlat
	log.Printf("Server starting on port %s in %s mode...", cfg.Server.Port, cfg.Server.Mode)
	router.Run(":" + cfg.Server.Port)
}
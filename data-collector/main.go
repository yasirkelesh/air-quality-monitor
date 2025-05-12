// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yasirkelesh/data-collector/api"
	"github.com/yasirkelesh/data-collector/config"
	"github.com/yasirkelesh/data-collector/messaging"
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
	log.Printf("MongoDB URI: %s, Database: %s, Collection: %s",
		cfg.MongoDB.URI,
		cfg.MongoDB.Database,
		cfg.MongoDB.Collection,
	)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer repo.Close()

	// RabbitMQ publisher oluştur
	rabbitConfig := messaging.RabbitMQConfig{
		URI:        cfg.RabbitMQ.URI,
		Exchange:   cfg.RabbitMQ.Exchange,
		Queue:      cfg.RabbitMQ.Queue,
		RoutingKey: cfg.RabbitMQ.RoutingKey,
	}

	var publisher messaging.MessagePublisher
	var rabbitErr error
	maxRetries := 10
	retryCount := 0

	for retryCount < maxRetries {
		publisher, rabbitErr = messaging.NewRabbitMQPublisher(rabbitConfig)
		if rabbitErr == nil {
			log.Printf("RabbitMQ bağlantısı başarılı: %s", cfg.RabbitMQ.URI)
			break
		}

		retryCount++
		log.Printf("RabbitMQ bağlantı denemesi %d/%d başarısız: %v", retryCount, maxRetries, rabbitErr)

		if retryCount < maxRetries {
			log.Printf("5 saniye sonra tekrar denenecek...")
			time.Sleep(5 * time.Second)
		}
	}

	if rabbitErr != nil {
		log.Fatalf("RabbitMQ bağlantısı %d deneme sonrasında başarısız oldu. Uygulama sonlandırılıyor.", maxRetries)
	}

	defer publisher.Close()

	// Servis katmanını oluştur
	pollutionService := service.NewPollutionService(repo, publisher)

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

	// HTTP sunucusu oluştur
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Ayrı bir goroutine'de sunucuyu başlat
	go func() {
		log.Printf("Server starting on port %s in %s mode...", cfg.Server.Port, cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown için sinyal yakalama
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Sunucuyu durdur
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

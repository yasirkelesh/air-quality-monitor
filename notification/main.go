package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"github.com/yasirkelesh/notification/api"
	"github.com/yasirkelesh/notification/config"
	"github.com/yasirkelesh/notification/domain"
	"github.com/yasirkelesh/notification/messaging"
	"github.com/yasirkelesh/notification/repository"
	"github.com/yasirkelesh/notification/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// Repository ve handler'ları oluştur
	userRepo, err := repository.NewUserRepository(
		cfg.MongoDB.URI,
		cfg.MongoDB.Database,
		cfg.MongoDB.Collection,
	)
	if err != nil {
		log.Fatalf("Failed to create user repository: %v", err)
	}
	defer userRepo.Close()

	rabbitConfig := messaging.RabbitMQConfig{
		URI:        cfg.RabbitMQ.URI,
		Exchange:   cfg.RabbitMQ.Exchange,
		Queue:      cfg.RabbitMQ.Queue,
		RoutingKey: cfg.RabbitMQ.RoutingKey,
	}

	log.Printf("RabbitMQ config: %+v", rabbitConfig)
	consumer, err := messaging.NewRabbitMQConsumer(rabbitConfig)
	if err != nil {
		log.Fatalf("Consumer başlatılamadı: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notificationService := service.NewNotificationService(
		&userRepo,
		cfg.Email.Host,
		cfg.Email.Port,
		cfg.Email.Username,
		cfg.Email.Password,
		cfg.Email.From,
	)

	// Anomali işleme

	go func() {
		err := consumer.Consume(ctx, func(d amqp.Delivery) {
			var anomaly domain.Anomaly

			if err := json.Unmarshal(d.Body, &anomaly); err != nil {
				log.Printf("JSON parse hatası: %v", err)
				return
			}

			log.Printf("Yeni mesaj alındı: %v", anomaly)
			err = notificationService.ProcessAnomaly(&anomaly)
			if err != nil {
				log.Printf("Anomali işlenemedi: %v", err)
			}
		})
		if err != nil {
			log.Fatalf("Consume başlatılamadı: %v", err)
		}
	}()

	/* userHandler := handlers.NewUserHandler(userRepo) */
	//gin modunu ayarla
	gin.SetMode(cfg.Server.Mode)
	// Gin router'ı oluştur
	r := gin.Default()

	// CORS middleware'i ekle
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api.SetupRoutes(r, &userRepo)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s in %s mode...", cfg.Server.Port, cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")

	/* // Health check endpoint'i
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Kullanıcı endpoint'leri
	r.POST("/user", userHandler.CreateUser)

	// Sunucuyu başlat
	r.Run(":8080") */
}

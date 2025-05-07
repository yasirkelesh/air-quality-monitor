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
	"github.com/yasirkelesh/notification/api"
	"github.com/yasirkelesh/notification/config"
	"github.com/yasirkelesh/notification/repository"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

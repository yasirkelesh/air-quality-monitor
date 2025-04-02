package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// Gin router oluştur
	router := gin.Default()

	// Basit bir endpoint ekle
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Sunucuyu başlat
	router.Run(":8080")
}

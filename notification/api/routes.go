package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yasirkelesh/notification/repository"
)

func SetupRoutes(router *gin.Engine, userRepo *repository.UserRepository) {
	userHandler := NewUserHandler(userRepo)

	//api/v1/users
	v1 := router.Group("/api/v1")
	{
		v1.POST("/users", userHandler.CreateUser)
		v1.GET("/health", userHandler.GetHealth)
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}

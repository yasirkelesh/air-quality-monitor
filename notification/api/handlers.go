package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yasirkelesh/notification/domain"
	"github.com/yasirkelesh/notification/repository"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: *userRepo,
	}
}

func (h *UserHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.userRepo.CreateUser(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Kullanıcı kaydedilemedi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kullanıcı başarıyla kaydedildi",
		"user":    user,
	})
}

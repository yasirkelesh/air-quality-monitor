// api/handlers.go (güncelleme)
package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yasirkelesh/data-collector/domain"
	"github.com/yasirkelesh/data-collector/service"
)

// PollutionHandler HTTP isteklerini işleyen yapı
type PollutionHandler struct {
	service *service.PollutionService
}

// NewPollutionHandler yeni bir handler oluşturur
func NewPollutionHandler(svc *service.PollutionService) *PollutionHandler {
	return &PollutionHandler{
		service: svc,
	}
}

// AddPollutionData veri ekleme endpoint'i
func (h *PollutionHandler) AddPollutionData(c *gin.Context) {
	var req PollutionDataRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  "error",
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Domain modeline dönüştür
	data := domain.NewPollutionData(req.Latitude, req.Longitude)
	data.PM25 = req.PM25
	data.PM10 = req.PM10
	data.NO2 = req.NO2
	data.SO2 = req.SO2
	data.O3 = req.O3

	// Timestamp varsa dönüştür
	if req.Timestamp != "" {
		if parsedTime, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			data.Timestamp = parsedTime
		}
	}

	// Device ID varsa ekle
	if req.DeviceID != "" {
		data.Source = req.DeviceID
	}

	// En az bir kirlilik verisi olmalı
	if !data.HasPollutants() {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Status:  "error",
			Error:   "missing_pollutants",
			Message: "At least one pollution parameter (PM2.5, PM10, NO2, SO2, O3) is required",
		})
		return
	}

	// Servis katmanını kullanarak verileri kaydet
	id, err := h.service.SavePollutionData(c.Request.Context(), data)
	if err != nil {
		log.Printf("Error saving pollution data: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  "error",
			Error:   "database_error",
			Message: "Failed to save pollution data",
		})
		return
	}

	// Başarılı yanıt döndür
	c.JSON(http.StatusCreated, PollutionDataResponse{
		Status:    "success",
		ID:        id,
		Message:   "Pollution data saved successfully",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// GetHealth sağlık kontrolü endpoint'i
func (h *PollutionHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "up",
		"service":   "data-collector",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "0.1.0",
	})
}

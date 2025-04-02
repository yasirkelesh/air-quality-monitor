// service/pollution_service.go
package service

import (
	"context"
	"log"
	"time"

	"github.com/yasirkelesh/data-collector/domain"
	"github.com/yasirkelesh/data-collector/mqtt"
	"github.com/yasirkelesh/data-collector/repository"
)

// PollutionService hava kirliliği verilerini işleyen servis
type PollutionService struct {
	repo       repository.PollutionRepository
	mqttClient *mqtt.Client
}

// NewPollutionService yeni bir servis oluşturur
func NewPollutionService(repo repository.PollutionRepository) *PollutionService {
	return &PollutionService{
		repo: repo,
	}
}

// InitMQTT MQTT istemcisini başlatır
func (s *PollutionService) InitMQTT(brokerURL, clientID, topic string) error {
	s.mqttClient = mqtt.NewClient(brokerURL, clientID, topic, s.HandleMQTTMessage)
	return s.mqttClient.Connect()
}

// CloseMQTT MQTT bağlantısını kapatır
func (s *PollutionService) CloseMQTT() {
	if s.mqttClient != nil {
		s.mqttClient.Disconnect()
	}
}

// SavePollutionData kirlilik verilerini kaydeder
func (s *PollutionService) SavePollutionData(ctx context.Context, data *domain.PollutionData) (string, error) {
	// Veri işleme/doğrulama işlemleri burada yapılabilir

	// Loglama
	log.Printf("Processing pollution data: lat=%f, lon=%f", data.Latitude, data.Longitude)

	// Zaman damgası yoksa ekle
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now().UTC()
	}

	// Repository'yi kullanarak verileri kaydet
	return s.repo.Save(ctx, data)
}

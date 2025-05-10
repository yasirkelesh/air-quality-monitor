// service/pollution_service.go
package service

import (
	"context"
	"log"
	"time"

	"github.com/yasirkelesh/data-collector/domain"
	"github.com/yasirkelesh/data-collector/messaging"
	"github.com/yasirkelesh/data-collector/mqtt"
	"github.com/yasirkelesh/data-collector/repository"
)


type PollutionService struct {
	repo             repository.PollutionRepository
	mqttClient       *mqtt.Client
	messagePublisher messaging.MessagePublisher
}

// NewPollutionService yeni bir servis oluşturur
func NewPollutionService(repo repository.PollutionRepository, publisher messaging.MessagePublisher) *PollutionService {
	return &PollutionService{
		repo:             repo,
		messagePublisher: publisher,
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

// SetMessagePublisher RabbitMQ mesaj yayınlayıcısını ayarlar
func (s *PollutionService) SavePollutionData(ctx context.Context, data *domain.PollutionData) (string, error) {
	// Veri işleme/doğrulama işlemleri burada yapılabilir
	
	// Zaman damgası yoksa ekle
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now().UTC()
	}

	// Repository'yi kullanarak verileri kaydet
	id, err := s.repo.Save(ctx, data)
	if err != nil {
		return "", err
	}

	// Kaydedilen ID'yi ekle
	data.ID = id

	// RabbitMQ'ya mesaj gönder
	if s.messagePublisher != nil { // Nil kontrolü eklendi
		routingKey := ""

		if err := s.messagePublisher.Publish(ctx, routingKey, data); err != nil {
			log.Printf("RabbitMQ'ya mesaj gönderme hatası: %v", err)
		} else {
			log.Printf("Pollution data published to RabbitMQ, ID: %s", id)
		}
	} else {
		log.Printf("RabbitMQ publisher is nil, skipping message publishing")
	}

	return id, nil
}

func (s *PollutionService) GetAllPollutionData(ctx context.Context, page, pageSize int) ([]*domain.PollutionData, int64, error) {
	// Verileri getir
	data, err := s.repo.FindAll(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Toplam kayıt sayısını getir (basit filtre, boş map)
	total, err := s.repo.CountData(ctx)
	if err != nil {
		return data, 0, err
	}

	return data, total, nil
}

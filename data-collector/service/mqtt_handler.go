// service/mqtt_handler.go
package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/yasirkelesh/data-collector/domain"
)

// MQTT üzerinden gelen verileri yapılandırma
type MQTTData struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Timestamp string   `json:"timestamp,omitempty"`
	PM25      *float64 `json:"pm25,omitempty"`
	PM10      *float64 `json:"pm10,omitempty"`
	NO2       *float64 `json:"no2,omitempty"`
	SO2       *float64 `json:"so2,omitempty"`
	O3        *float64 `json:"o3,omitempty"`
	DeviceID  string   `json:"device_id,omitempty"`
}

// HandleMQTTMessage MQTT mesajlarını işler ve servise yönlendirir
func (s *PollutionService) HandleMQTTMessage(client mqtt.Client, msg mqtt.Message) {
	log.Printf("MQTT Mesajı alındı - Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())

	// JSON verisini ayrıştır
	var mqttData MQTTData
	if err := json.Unmarshal(msg.Payload(), &mqttData); err != nil {
		log.Printf("MQTT mesajı JSON olarak ayrıştırılamadı: %v\n", err)
		return
	}

	// Domain modeline dönüştür
	data := domain.NewPollutionData(mqttData.Latitude, mqttData.Longitude)
	data.PM25 = mqttData.PM25
	data.PM10 = mqttData.PM10
	data.NO2 = mqttData.NO2
	data.SO2 = mqttData.SO2
	data.O3 = mqttData.O3
	data.Source = "mqtt"

	// Device ID varsa ekle
	if mqttData.DeviceID != "" {
		data.Source = mqttData.DeviceID
	}

	// Timestamp varsa parse et
	if mqttData.Timestamp != "" {
		if parsedTime, err := time.Parse(time.RFC3339, mqttData.Timestamp); err == nil {
			data.Timestamp = parsedTime
		}
	}

	// En az bir kirlilik verisi olmalı
	if !data.HasPollutants() {
		log.Println("MQTT mesajında geçerli kirlilik parametresi yok")
		return
	}

	// Veriyi kaydet
	id, err := s.SavePollutionData(context.Background(), data)
	if err != nil {
		log.Printf("MQTT verisi kaydedilirken hata: %v", err)
		return
	}

	log.Printf("MQTT verisi başarıyla kaydedildi, ID: %s", id)
}

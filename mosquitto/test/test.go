package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker      = "tcp://localhost:1883" // MQTT broker adresini güncelle
	topic       = "pollution"          // Yayınlanacak topic
	clientID    = "go_mqtt_publisher"
	interval    = 1 * time.Second // Veri gönderme aralığı
	maxSensorID = 5               // Kaç farklı sensör ID'si olacak
)

type SensorData struct {
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

// RandomFloat belirtilen aralıkta rastgele float değer üretir
func RandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func main() {
	// Rastgele sayı üreteci için seed
	rand.Seed(time.Now().UnixNano())

	// MQTT bağlantı ayarları
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	// Bağlantı yapılandırması
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("MQTT broker'a bağlandı. Veri göndermeye başlıyor...")
	fmt.Printf("Veri gönderme aralığı: %s\n", interval)

	// Döngü içinde sürekli veri gönder
	counter := 0
	for {
		counter++
		// Rastgele bir sensör ID'si seç
		sensorID := fmt.Sprintf("temp_sensor_%02d", rand.Intn(maxSensorID)+1)

		// Sensör verileri oluştur
		data := SensorData{
			//Latitude:  RandomFloat(40.7128, 41.94),
			//Longitude: RandomFloat(29.949065, 30.224995),
			Latitude: 40.714331,
			Longitude: 29.945292,
			Timestamp: time.Now().Format(time.RFC3339),
			DeviceID:  sensorID,
		}
		data.PM25 = new(float64)
		*data.PM25 = RandomFloat(0, 10)
		data.PM10 = new(float64)
		*data.PM10 = RandomFloat(0, 10)
		data.NO2 = new(float64)
		*data.NO2 = RandomFloat(0, 10)
		data.SO2 = new(float64)
		*data.SO2 = RandomFloat(0, 10)
		data.O3 = new(float64)
		*data.O3 = RandomFloat(0, 10)
		// JSON formatına dönüştür
		payload, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("JSON oluşturma hatası: %v\n", err)
			continue
		}

		// MQTT üzerinden yayınla
		token := client.Publish(topic, 0, false, payload)
		token.Wait()

		fmt.Printf("Veri gönderildi: %s\n", payload)

		// Belirtilen süre kadar bekle
		time.Sleep(interval)
	}

	// Not: Bu koda asla ulaşılmayacak, çünkü sonsuz döngü var
	// client.Disconnect(250)
}

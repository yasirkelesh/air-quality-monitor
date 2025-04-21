package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker   = "tcp://localhost:1883"
	topic    = "pollution"
	clientID = "go_mqtt_publisher"
	interval = 1 * time.Second
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

func RandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func getTestScenario() int {
	var choice int
	fmt.Println("Hangi test senaryosu uygulanacak?")
	fmt.Println("1 - Normal veri gönderimi")
	fmt.Println("2 - Anomali testi")
	fmt.Println("3 - Türkiye'deki 40 farklı lokasyondan veri gönderimi")
	fmt.Print("Seçiminiz: ")
	fmt.Scan(&choice)
	return choice
}

func connectMQTT() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("MQTT broker'a bağlandı.")
	return client
}

func publishData(client mqtt.Client, data SensorData) {
	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("JSON oluşturma hatası: %v\n", err)
		return
	}
	token := client.Publish(topic, 0, false, payload)
	token.Wait()
	fmt.Printf("Veri gönderildi: %s\n", payload)
}

func normalDataTest(client mqtt.Client) {
	for {
		data := generateSensorData("sensor_normal", 40.714331, 29.945292, false)
		publishData(client, data)
		time.Sleep(interval)
	}
}

func anomalyTest(client mqtt.Client) {
	counter := 0
	for {
		anomali := counter >= 10 // İlk 10 veri normal, sonra anomaliler başlar
		data := generateSensorData("sensor_anomaly", 40.714331, 29.945292, anomali)
		publishData(client, data)
		counter++
		time.Sleep(interval)
	}
}

func multiLocationTest(client mqtt.Client) {
	// Türkiye'den rastgele 40 lokasyon
	locations := generateRandomLocations(40)
	for {
		i := rand.Intn(len(locations))
		lat := locations[i][0]
		lon := locations[i][1]
		deviceID := fmt.Sprintf("sensor_tr_%02d", i+1)
		data := generateSensorData(deviceID, lat, lon, false)
		publishData(client, data)
		time.Sleep(interval)
	}
}

func generateSensorData(deviceID string, lat, lon float64, isAnomaly bool) SensorData {
	data := SensorData{
		Latitude:  lat,
		Longitude: lon,
		Timestamp: time.Now().Format(time.RFC3339),
		DeviceID:  deviceID,
	}
	data.PM25 = new(float64)
	data.PM10 = new(float64)
	data.NO2 = new(float64)
	data.SO2 = new(float64)
	data.O3 = new(float64)

	if isAnomaly {
		*data.PM25 = RandomFloat(500, 1000)
		*data.PM10 = RandomFloat(400, 800)
		*data.NO2 = RandomFloat(200, 600)
		*data.SO2 = RandomFloat(100, 400)
		*data.O3 = RandomFloat(300, 700)
	} else {
		*data.PM25 = RandomFloat(0, 50)
		*data.PM10 = RandomFloat(0, 50)
		*data.NO2 = RandomFloat(0, 50)
		*data.SO2 = RandomFloat(0, 50)
		*data.O3 = RandomFloat(0, 50)
	}

	return data
}

func generateRandomLocations(count int) [][]float64 {
	var locations [][]float64
	for i := 0; i < count; i++ {
		lat := RandomFloat(36.0, 42.0) // Türkiye enlemleri
		lon := RandomFloat(26.0, 45.0) // Türkiye boylamları
		locations = append(locations, []float64{lat, lon})
	}
	return locations
}

func main() {
	rand.Seed(time.Now().UnixNano())
	scenario := getTestScenario()
	client := connectMQTT()

	switch scenario {
	case 1:
		normalDataTest(client)
	case 2:
		anomalyTest(client)
	case 3:
		multiLocationTest(client)
	default:
		fmt.Println("Geçersiz seçim. Program sonlandırılıyor.")
	}
}

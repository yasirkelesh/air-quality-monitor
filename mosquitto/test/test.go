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
		data := generateSensorData("sensor_normal", 40.714331, 29.945292, false, 0)
		publishData(client, data)
		time.Sleep(interval)
	}
}

func anomalyTest(client mqtt.Client) {
	lat := 40.714331
	lon := 29.945292
	counter := 0

	for {
		counter++
		deviceID := "sensor_anomaly_01"
		data := generateSensorData(deviceID, lat, lon, true, counter)
		publishData(client, data)
		time.Sleep(interval)
	}
}


func multiLocationTest(client mqtt.Client) {
	locations := getFixedLocations()
	index := 0
	counter := 0

	for {
		lat := locations[index][0]
		lon := locations[index][1]
		deviceID := fmt.Sprintf("sensor_tr_%02d", index+1)
		data := generateSensorData(deviceID, lat, lon, false, counter)
		publishData(client, data)

		index = (index + 1) % len(locations)
		counter++
		time.Sleep(interval)
	}
}

func generateSensorData(deviceID string, lat, lon float64, isAnomaly bool, counter int) SensorData {
	base := 25.0

	var pm25, pm10, no2, so2, o3 float64

	// Normal sabit değerler
	pm25, pm10, no2, so2, o3 = base, base+1, base-1, base, base+0.5

	// Eğer anomali oluşturulacaksa bazı değerleri aşırı değiştir
	if isAnomaly && counter%10 == 0 {
		pm25 += RandomFloat(50, 100)  // örnek: 75 gibi yüksek bir değer
		pm10 += RandomFloat(40, 90)
	}

	return SensorData{
		Latitude:  lat,
		Longitude: lon,
		Timestamp: time.Now().Format(time.RFC3339),
		DeviceID:  deviceID,
		PM25:      &pm25,
		PM10:      &pm10,
		NO2:       &no2,
		SO2:       &so2,
		O3:        &o3,
	}
}


// Türkiye'deki sabit 40 lokasyon (şehir merkezleri)
func getFixedLocations() [][]float64 {
	return [][]float64{
		{41.0082, 28.9784},  // İstanbul
		{39.9208, 32.8541},  // Ankara
		{38.4192, 27.1287},  // İzmir
		{37.0662, 37.3833},  // Gaziantep
		{36.8969, 30.7133},  // Antalya
		{40.1828, 29.0663},  // Bursa
		{37.8746, 32.4932},  // Konya
		{37.0031, 35.3213},  // Adana
		{38.6743, 39.2232},  // Elazığ
		{38.3552, 38.3095},  // Malatya
		{39.7477, 37.0179},  // Sivas
		{37.7648, 30.5566},  // Isparta
		{41.2867, 36.33},    // Samsun
		{40.6500, 35.8333},  // Çorum
		{38.4622, 27.2164},  // Manisa
		{39.7191, 43.0519},  // Ağrı
		{37.8560, 40.5350},  // Diyarbakır
		{36.8121, 34.6415},  // Mersin
		{38.7322, 35.4853},  // Kayseri
		{38.6823, 34.8554},  // Nevşehir
		{37.2156, 28.3636},  // Muğla
		{38.2456, 29.4082},  // Uşak
		{38.6785, 43.3826},  // Van
		{37.7694, 38.2786},  // Adıyaman
		{38.4756, 43.3790},  // Bitlis
		{37.5526, 36.9371},  // Kahramanmaraş
		{40.9833, 37.8667},  // Ordu
		{37.5183, 36.9351},  // Osmaniye
		{39.9097, 41.2753},  // Erzurum
		{40.5760, 31.5800},  // Bolu
		{41.0015, 40.5124},  // Rize
		{41.0053, 29.0123},  // İstanbul (tekrar, Asya yakası)
		{38.4967, 42.2816},  // Muş
		{39.9096, 41.2746},  // Erzurum (merkez tekrar)
		{37.0738, 37.3834},  // Gaziantep (detay)
		{37.7510, 29.0870},  // Denizli
		{38.0933, 36.7125},  // Tufanbeyli (Adana)
		{39.1960, 27.1812},  // Balıkesir
		{38.5019, 39.5269},  // Bingöl
	}
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

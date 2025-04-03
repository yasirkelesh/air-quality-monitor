// config/config.go
package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config uygulama konfigürasyonu
type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	MQTT     MQTTConfig
	RabbitMQ RabbitMQConfig
}

// ServerConfig HTTP sunucu ayarları
type ServerConfig struct {
	Port string
	Mode string // debug, release, test
}

// MongoDBConfig MongoDB bağlantı ayarları
type MongoDBConfig struct {
	URI        string
	Database   string
	Collection string
}

// MQTTConfig MQTT bağlantı ayarları
type MQTTConfig struct {
	BrokerURL string
	ClientID  string
	Topic     string
}

type RabbitMQConfig struct {
	URI      string
	Exchange string
	Queue    string
	RoutingKey string
}


// LoadConfig konfigürasyon yükler
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config") // config.yaml, config.json, vb.
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv() // Çevresel değişkenleri de oku

	// Varsayılan değerleri ayarla
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "pollution_db")
	viper.SetDefault("mongodb.collection", "raw_data")
	
	viper.SetDefault("mqtt.brokerurl", "mqtt://localhost:1883")
	viper.SetDefault("mqtt.clientid", "data-collector")
	viper.SetDefault("mqtt.topic", "pollution/#")

	viper.SetDefault("rabbitmq.uri", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("rabbitmq.exchange", "pollution.data")
	viper.SetDefault("rabbitmq.queue", "raw-data")
	viper.SetDefault("rabbitmq.routingkey", "raw.data")

	// Konfigürasyon dosyasını oku
	if err := viper.ReadInConfig(); err != nil {
		// Konfigürasyon dosyası yoksa uyarı ver, varsayılan değerleri kullan
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found. Using default values.")
		} else {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
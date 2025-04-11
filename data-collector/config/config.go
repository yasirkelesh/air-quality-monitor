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
	URI        string `mapstructure:"uri"`
	Database   string `mapstructure:"database"`
	Collection string `mapstructure:"collection"`
}

// MQTTConfig MQTT bağlantı ayarları
type MQTTConfig struct {
	BrokerURL string `mapstructure:"brokerurl"`
	ClientID  string `mapstructure:"clientid"`
	Topic     string `mapstructure:"topic"`
}

type RabbitMQConfig struct {
	URI        string `mapstructure:"uri"`
	Exchange   string `mapstructure:"exchange"`
	Queue      string `mapstructure:"queue"`
	RoutingKey string `mapstructure:"routingkey"`
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

	viper.SetDefault("mongodb.uri", "mongodb://mongodb:27017")
	viper.SetDefault("mongodb.database", "pollution_db")
	viper.SetDefault("mongodb.collection", "raw_data")

	viper.SetDefault("mqtt.brokerurl", "mqtt://mqtt-broker:1883")
	viper.SetDefault("mqtt.clientid", "data-collector")
	viper.SetDefault("mqtt.topic", "pollution/#")

	viper.SetDefault("rabbitmq.uri", "amqp://guest:guest@rabbitmqt:5672/") //buna tekrar bak gust:guest@rabbitmq:5672
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

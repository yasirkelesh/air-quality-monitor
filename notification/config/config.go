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
	RabbitMQ RabbitMQConfig
	Email    EmailConfig
}

// ServerConfig HTTP sunucu ayarları
type ServerConfig struct {
	Port string
	Mode string // debug, release, test
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// MongoDBConfig MongoDB bağlantı ayarları
type MongoDBConfig struct {
	URI        string `mapstructure:"uri"`
	Database   string `mapstructure:"database"`
	Collection string `mapstructure:"collection"`
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

	// Varsayılan değerleri ayarla ayarlamasan da olur
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("mongodb.uri", "mongodb://mongodb:27017")
	viper.SetDefault("mongodb.database", "notification_db")
	viper.SetDefault("mongodb.collection", "users")

	viper.SetDefault("rabbitmq.uri", "amqp://guest:guest@rabbitmqt:5672/") //buna tekrar bak gust:guest@rabbitmq:5672
	viper.SetDefault("rabbitmq.exchange", "pollution.data")
	viper.SetDefault("rabbitmq.queue", "anomaly-data")
	viper.SetDefault("rabbitmq.routingkey", "anomaly.data")

	viper.SetDefault("email.host", "smtp.gmail.com")
	viper.SetDefault("email.port", 587)
	viper.SetDefault("email.username", "your-email@gmail.com")
	viper.SetDefault("email.password", "your-email-password")
	viper.SetDefault("email.from", "your-email@gmail.com")

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

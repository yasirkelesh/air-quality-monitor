// config/config.go
package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config API Gateway konfigürasyon yapısı
type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	Auth     AuthConfig
}

// ServerConfig HTTP sunucu ayarları
type ServerConfig struct {
	Port string
	Mode string
}

// ServicesConfig mikroservis bağlantı adresleri
type ServicesConfig struct {
	DataCollector    string
	DataProcessing   string
	Notification     string
	AnomalyDetection string
	// İleride eklenecek diğer servisler burada tanımlanabilir
}

// AuthConfig kimlik doğrulama ayarları
type AuthConfig struct {
	JWTSecret string
	Enabled   bool
}

// LoadConfig konfigürasyon yükleme
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	// Varsayılan değerleri ayarla
	viper.SetDefault("server.port", "8000")
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("services.datacollector", "http://data-collector:8080")
	viper.SetDefault("services.dataprocessing", "http://data-processing:5000")
	viper.SetDefault("services.notification", "http://notification:9090")
	viper.SetDefault("services.anomalydetection", "http://anomaly-detection:6000")

	viper.SetDefault("auth.jwtsecret", "your-secret-key")
	viper.SetDefault("auth.enabled", false)

	if err := viper.ReadInConfig(); err != nil {
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

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

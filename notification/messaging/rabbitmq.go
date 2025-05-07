// messaging/rabbitmq.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// MessagePublisher mesaj yayınlama arayüzü
type MessagePublisher interface {
	Publish(ctx context.Context, routingKey string, data interface{}) error
	Close() error
}

// RabbitMQConfig RabbitMQ yapılandırması
type RabbitMQConfig struct {
	URI        string
	Exchange   string
	Queue      string
	RoutingKey string
}

// RabbitMQPublisher RabbitMQ ile iletişim kuran yapı
type RabbitMQPublisher struct {
	config     RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewRabbitMQPublisher yeni bir RabbitMQ publisher oluşturur
func NewRabbitMQPublisher(config RabbitMQConfig) (*RabbitMQPublisher, error) {
	// RabbitMQ'ya bağlan
	conn, err := amqp.Dial(config.URI)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQ bağlantı hatası: %v", err)
	}

	// Kanal oluştur
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("RabbitMQ kanal hatası: %v", err)
	}

	// Exchange oluştur
	err = ch.ExchangeDeclare(
		config.Exchange, // exchange adı
		"topic",         // exchange tipi
		true,            // dayanıklı (durable)
		false,           // otomatik silme (auto-delete)
		false,           // dahili (internal)
		false,           // bekletme yok (no-wait)
		nil,             // argümanlar
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("RabbitMQ exchange oluşturma hatası: %v", err)
	}

	// Kuyruk oluştur
	_, err = ch.QueueDeclare(
		config.Queue, // kuyruk adı
		true,         // dayanıklı (durable)
		false,        // otomatik silme (auto-delete)
		false,        // özel (exclusive)
		false,        // bekletme yok (no-wait)
		nil,          // argümanlar
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("RabbitMQ kuyruk oluşturma hatası: %v", err)
	}

	// Kuyruk ile exchange'i bağla
	err = ch.QueueBind(
		config.Queue,      // kuyruk adı
		config.RoutingKey, // routing key
		config.Exchange,   // exchange adı
		false,             // bekletme yok (no-wait)
		nil,               // argümanlar
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("RabbitMQ kuyruk bağlama hatası: %v", err)
	}

	log.Printf("RabbitMQ bağlantısı başarılı: %s\n", config.URI)
	return &RabbitMQPublisher{
		config:     config,
		connection: conn,
		channel:    ch,
	}, nil
}

// Publish verilen veriyi RabbitMQ'ya gönderir
func (p *RabbitMQPublisher) Publish(ctx context.Context, routingKey string, data interface{}) error {
	// Routing key belirtilmemişse varsayılanı kullan
	if routingKey == "" {
		routingKey = p.config.RoutingKey
	}

	// Veriyi JSON'a dönüştür
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("JSON dönüştürme hatası: %v", err)
	}

	// RabbitMQ'ya gönder
	err = p.channel.Publish(
		p.config.Exchange, // exchange
		routingKey,        // routing key
		false,             // zorunlu (mandatory)
		false,             // acil (immediate)
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Mesajın kalıcı olmasını sağlar
		})
	if err != nil {
		return fmt.Errorf("RabbitMQ mesaj gönderme hatası: %v", err)
	}

	return nil
}

// Close RabbitMQ bağlantısını kapatır
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		return p.connection.Close()
	}
	return nil
}

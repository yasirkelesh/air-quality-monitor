// messaging/rabbitmq.go
package messaging

import (
	"context"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// MessagePublisher mesaj yayınlama arayüzü
type MessageConsumer interface {
	Consume(ctx context.Context, handler func(d amqp.Delivery)) error
	Close() error
}

// RabbitMQConfig RabbitMQ yapılandırması
type RabbitMQConfig struct {
	URI        string
	Exchange   string
	Queue      string
	RoutingKey string
}

type RabbitMQConsumer struct {
	config     RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewRabbitMQPublisher yeni bir RabbitMQ publisher oluşturur
func NewRabbitMQConsumer(config RabbitMQConfig) (*RabbitMQConsumer, error) {
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

	log.Printf("RabbitMQ bağlantısı başarılı: %s\n", config.URI)

	return &RabbitMQConsumer{
		config:     config,
		connection: conn,
		channel:    ch,
	}, nil
}

// Consume mesajları tüketir ve her mesaj için handler fonksiyonunu çağırır
func (c *RabbitMQConsumer) Consume(ctx context.Context, handler func(d amqp.Delivery)) error {
	// Exchange ve queue'yu tanımla
	err := c.channel.ExchangeDeclare(
		c.config.Exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("Exchange oluşturulamadı: %v", err)
	}

	_, err = c.channel.QueueDeclare(
		c.config.Queue, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue oluşturulamadı: %v", err)
	}

	err = c.channel.QueueBind(
		c.config.Queue,      // queue name
		c.config.RoutingKey, // routing key
		c.config.Exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Queue bind hatası: %v", err)
	}

	msgs, err := c.channel.Consume(
		c.config.Queue, // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		return fmt.Errorf("Mesajlar alınamadı: %v", err)
	}

	// Mesajları async olarak işle
	go func() {
		for {
			select {
			case msg := <-msgs:
				handler(msg)
			case <-ctx.Done():
				log.Println("Consumer durduruldu.")
				return
			}
		}
	}()

	return nil
}

// Close RabbitMQ bağlantısını kapatır
func (p *RabbitMQConsumer) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		return p.connection.Close()
	}
	return nil
}

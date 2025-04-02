// mqtt/client.go
package mqtt

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Client MQTT istemcisi
type Client struct {
	client     mqtt.Client
	brokerURL  string
	clientID   string
	topic      string
	messageHandler mqtt.MessageHandler
}

// NewClient yeni bir MQTT istemcisi oluşturur
func NewClient(brokerURL, clientID, topic string, messageHandler mqtt.MessageHandler) *Client {
	return &Client{
		brokerURL:      brokerURL,
		clientID:       clientID,
		topic:          topic,
		messageHandler: messageHandler,
	}
}

// Connect MQTT broker'a bağlanır
func (c *Client) Connect() error {
	// MQTT bağlantı seçenekleri
	opts := mqtt.NewClientOptions().
		AddBroker(c.brokerURL).
		SetClientID(c.clientID).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectTimeout(5 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetDefaultPublishHandler(c.messageHandler)

	// OnConnect callback - bağlantı kurulduğunda yapılacak işler
	opts.OnConnect = func(client mqtt.Client) {
		// Topic'e abone ol
		if token := client.Subscribe(c.topic, 0, c.messageHandler); token.Wait() && token.Error() != nil {
			log.Printf("Topic'e abone olunurken hata oluştu: %v\n", token.Error())
		} else {
			log.Printf("MQTT broker'a bağlandı ve %s topic'ine abone oldu\n", c.topic)
		}
	}

	// MQTT istemcisini oluştur
	c.client = mqtt.NewClient(opts)

	// Broker'a bağlan
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("MQTT broker'a bağlanırken hata: %v", token.Error())
	}

	return nil
}

// Disconnect MQTT bağlantısını kapatır
func (c *Client) Disconnect() {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(250)
	}
}

// IsConnected istemcinin bağlı olup olmadığını kontrol eder
func (c *Client) IsConnected() bool {
	return c.client != nil && c.client.IsConnected()
}
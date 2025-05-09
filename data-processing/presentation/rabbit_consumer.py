import json
import pika
from loguru import logger
from typing import Callable

from config import (
    RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_USER,
    RABBITMQ_PASS, RABBITMQ_RAW_QUEUE,
    RABBITMQ_RAW_ROUTING_KEY,
)

class RabbitMQConsumer:
    """RabbitMQ'dan ham veri almak için consumer sınıfı"""
    
    def __init__(self, callback: Callable):
        """
        RabbitMQ Consumer başlat
        
        Args:
            callback: Ham veri alındığında çağrılacak fonksiyon
        """
        self.callback = callback
        self.connection = None
        self.channel = None
        
    def connect(self):
        """RabbitMQ'ya bağlan"""
        try:
            # Bağlantı parametrelerini ayarla
            credentials = pika.PlainCredentials(RABBITMQ_USER, RABBITMQ_PASS)
            parameters = pika.ConnectionParameters(
                host=RABBITMQ_HOST,
                port=RABBITMQ_PORT,
                credentials=credentials
            )
            
            # Bağlantı ve kanal aç
            self.connection = pika.BlockingConnection(parameters)
            self.channel = self.connection.channel()
            
            # Kuyruğu tanımla
            self.channel.queue_declare(queue=RABBITMQ_RAW_QUEUE, durable=True)
            
            logger.info(f"RabbitMQ'ya bağlantı başarılı: {RABBITMQ_HOST}:{RABBITMQ_PORT}")
            return True
            
        except Exception as e:
            logger.error(f"RabbitMQ bağlantı hatası: {str(e)}")
            return False
    
    def start_consuming(self):
        """Mesaj almaya başla"""
        try:
            def process_message(ch, method, properties, body):
                try:
                    # Mesajı JSON'a dönüştür
                    message = json.loads(body)
                    logger.info(f"Yeni mesaj alındı: {message.get('source', 'unknown')}")
                    
                    # Callback fonksiyonu çağır
                    self.callback(message)
                    
                    # Mesajı onaylayarak kuyruktan çıkar
                    ch.basic_ack(delivery_tag=method.delivery_tag)
                
                except json.JSONDecodeError:
                    logger.error("Geçersiz JSON verisi")
                    ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
                
                except Exception as e:
                    logger.error(f"Mesaj işleme hatası: {str(e)}")
                    ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)
            
            # Mesaj işleme fonksiyonunu ayarla
            self.channel.basic_consume(
                queue=RABBITMQ_RAW_QUEUE,
                on_message_callback=process_message
            )
            
            # Mesaj beklemeye başla
            logger.info(f"{RABBITMQ_RAW_QUEUE} kuyruğundan mesaj bekleniyor...")
            self.channel.start_consuming()
        
        except Exception as e:
            logger.error(f"Consumer hatası: {str(e)}")
    
    def close(self):
        """Bağlantıyı kapat"""
        if self.connection and self.connection.is_open:
            self.connection.close()
            logger.info("RabbitMQ bağlantısı kapatıldı")
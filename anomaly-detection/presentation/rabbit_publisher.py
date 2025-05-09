import json
import pika
from loguru import logger
from datetime import datetime

from config import (
    RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_USER,
    RABBITMQ_PASS, RABBITMQ_ANOMALY_QUEUE,
    RABBITMQ_ANOMALY_ROUTING_KEY,
    RABBITMQ_EXCHANGE,
)
class DateTimeEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, datetime):
            return obj.isoformat()
        return super(DateTimeEncoder, self).default(obj)
    
class RabbitMQPublisher:
    """İşlenmiş verileri RabbitMQ'ya göndermek için publisher sınıfı"""
    
    def __init__(self):
        """RabbitMQ Publisher başlat"""
        self.connection = None
        self.channel = None
        self.connect()
        
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
           
            # Exchange'i tanımla
            self.channel.exchange_declare(
                exchange=RABBITMQ_EXCHANGE,
                exchange_type='topic',
                durable=True
            )

            # Kuyruğu tanımla
            self.channel.queue_declare(queue=RABBITMQ_ANOMALY_QUEUE, durable=True)
            
            # Queue'yu exchange'e bağla
            self.channel.queue_bind(
                queue=RABBITMQ_ANOMALY_QUEUE,
                exchange=RABBITMQ_EXCHANGE,
                routing_key=RABBITMQ_ANOMALY_ROUTING_KEY
            )
            
            logger.info(f"RabbitMQ Publisher bağlantı başarılı: {RABBITMQ_HOST}:{RABBITMQ_PORT}")
            logger.info(f"RabbitMQ Publisher kuyruk başarılı: {RABBITMQ_ANOMALY_QUEUE}")
            return True
            
        except Exception as e:
            logger.error(f"RabbitMQ Publisher bağlantı hatası: {str(e)}")
            return False

    def publish(self, data):
        """İşlenmiş veriyi kuyruğa gönder"""
        try:
            # Veriyi JSON'a dönüştür
            message = json.dumps(data, cls=DateTimeEncoder)
            
            # Kuyruğa gönder
            self.channel.basic_publish(
                exchange=RABBITMQ_EXCHANGE,
                routing_key=RABBITMQ_ANOMALY_ROUTING_KEY,
                body=message,
                properties=pika.BasicProperties(
                    delivery_mode=2,  # Kalıcı mesaj
                    content_type='application/json'
                )
            )
            
            logger.info(f"Mesaj başarıyla gönderildi: {data.get('source', 'unknown')}")
            return True
            
        except Exception as e:
            logger.error(f"Mesaj gönderme hatası: {str(e)}")
            
            # Bağlantı kopmuşsa yeniden bağlanmayı dene
            if not self.connection or self.connection.is_closed:
                logger.info("Bağlantı kopmuş, yeniden bağlanmayı deniyorum...")
                self.connect()
                
            return False
    
    def close(self):
        """Bağlantıyı kapat"""
        if self.connection and self.connection.is_open:
            self.connection.close()
            logger.info("RabbitMQ Publisher bağlantısı kapatıldı")
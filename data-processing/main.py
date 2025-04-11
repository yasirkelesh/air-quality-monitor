import threading
import time
import uvicorn
from loguru import logger

from config import API_HOST, API_PORT
from presentation.rabbit_consumer import RabbitMQConsumer
from presentation.rabbit_publisher import RabbitMQPublisher

# Geçici veri işleme fonksiyonu
def process_data(raw_data):
    """
    Geçici veri işleme fonksiyonu - ileride gerçek implementasyon ile değiştirilecek
    """
    logger.info(f"Veri işleniyor: {raw_data}")
    
    # Basit bir işleme: ham veriyi alıp, işlenmiş olarak işaretle
    processed_data = raw_data.copy()
    processed_data['processed'] = True
    processed_data['processed_at'] = time.strftime('%Y-%m-%dT%H:%M:%SZ')
    
    # İşlenmiş veriyi gönder
    publisher = RabbitMQPublisher()
    publisher.publish(processed_data)
    publisher.close()

def start_api_server():
    """API sunucusunu başlat"""
    logger.info("API sunucusu başlatılıyor...")
    from presentation.api.app import app
    uvicorn.run(app, host=API_HOST, port=API_PORT)

def start_rabbitmq_consumer():
    """RabbitMQ consumer'ı başlat"""
    logger.info("RabbitMQ consumer başlatılıyor...")
    consumer = RabbitMQConsumer(callback=process_data)
    
    # RabbitMQ'ya bağlan
    if consumer.connect():
        # Mesaj almaya başla
        consumer.start_consuming()
    else:
        logger.error("RabbitMQ bağlantısı kurulamadı, yeniden deneniyor...")
        time.sleep(5)  # 5 saniye bekle
        start_rabbitmq_consumer()  # Recursion ile yeniden dene

if __name__ == "__main__":
    logger.info("Veri İşleme Servisi başlatılıyor...")
    
    # API'yi ayrı bir thread'de başlat
    api_thread = threading.Thread(target=start_api_server)
    api_thread.daemon = True
    api_thread.start()
    
    # RabbitMQ consumer'ı ana thread'de çalıştır
    start_rabbitmq_consumer()
import threading
import time
import uvicorn
from loguru import logger

from config import API_HOST, API_PORT
from presentation.rabbit_consumer import RabbitMQConsumer
from presentation.rabbit_publisher import RabbitMQPublisher
from business.data_processor import DataProcessor
from business.aggregation_service import AggregationService
from data_access.influxdb_repository import InfluxDBRepository

# Geçici veri işleme fonksiyonu
def process_data(raw_data):
    """
    RabbitMQ'dan gelen ham veriyi işle ve sonucu gönder
    """
    try:
        logger.info(f"Veri işleniyor: {raw_data.get('source', 'unknown')}")
        
        # Veri işleme servisi ile veriyi işle
        processor = DataProcessor()
        processed_data = processor.process(raw_data)
        
        # Bölgesel ortalama hesapla ve ekle
        aggregation_service = AggregationService()
        regional_avg = aggregation_service.get_regional_average(processed_data.get("geohash"))
        
        if regional_avg:
            processed_data["regional_average"] = regional_avg
        
        # İşlenmiş veriyi kontrol et
        logger.info(f"İşlenmiş veri: {processed_data}")
        # İşlenmiş veriyi gönder
        publisher = RabbitMQPublisher()
        publisher.publish(processed_data)
        publisher.close()
        
        logger.info(f"Veri işlendi ve gönderildi: {processed_data.get('source', 'unknown')}")
        
    except Exception as e:
        logger.error(f"Veri işleme hatası: {str(e)}")

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
# main.py
import time
from loguru import logger
from presentation.rabbit_consumer import RabbitMQConsumer
from business.anomaly_detector import AnomalyDetector
from domain.models.processed_data import ProcessedData

def process_message(message):
    """
    RabbitMQ'dan gelen mesajı işle
    
    Args:
        message: JSON mesajı
    """
    try:
        # Veriyi model nesnesine dönüştür
        data = ProcessedData(message)
        
        logger.info(f"İşlenmiş veri: {message}")
        # Anomali tespiti yap
        detector = AnomalyDetector()
        anomalies = detector.detect(data)
        
        # Tespit edilen anomalileri yazdır
        if anomalies:
            logger.warning(f"Anomali tespit edildi - Kaynak: {data.source}")
            for anomaly in anomalies:
                logger.warning(f"[{anomaly.anomaly_type}] {anomaly.description}")
                # TODO: MongoDB'ye kaydet
                # TODO: WebSocket üzerinden bildir
        else:
            logger.info(f"Anomali bulunamadı - Kaynak: {data.source}")
    
    except Exception as e:
        logger.error(f"Mesaj işleme hatası: {str(e)}")

def start_rabbitmq_consumer():
    """RabbitMQ consumer'ı başlat"""
    consumer = RabbitMQConsumer(callback=process_message)
    
    # RabbitMQ'ya bağlan
    if consumer.connect():
        # Mesaj almaya başla
        consumer.start_consuming()
    else:
        logger.error("RabbitMQ bağlantısı kurulamadı, yeniden deneniyor...")
        time.sleep(5)
        start_rabbitmq_consumer()

if __name__ == "__main__":
    logger.info("Anomali Tespiti Servisi başlatılıyor...")
    
    # RabbitMQ consumer'ı başlat
    start_rabbitmq_consumer()
# main.py - SSE entegrasyonu ile güncellenmiş versiyon
import time
import threading
import asyncio
from loguru import logger
from presentation.rabbit_consumer import RabbitMQConsumer
from presentation.sse_controller import SSEController
from business.anomaly_detector import AnomalyDetector
from domain.models.processed_data import ProcessedData
from data_access.anomaly_repository import AnomalyRepository

# Repository ve SSE controller başlat
anomaly_repository = AnomalyRepository()
sse_controller = SSEController()

def process_message(message):
    """
    RabbitMQ'dan gelen mesajı işle
    
    Args:
        message: JSON mesajı
    """
    try:
        # Veriyi model nesnesine dönüştür
        data = ProcessedData(message)
        
        # Anomali tespiti yap
        detector = AnomalyDetector()
        anomalies = detector.detect(data)
        
        # Tespit edilen anomalileri kaydet ve bildir
        if anomalies:
            logger.warning(f"Anomali tespit edildi - Kaynak: {data.source}")
            for anomaly in anomalies:
                logger.warning(f"[{anomaly.anomaly_type}] {anomaly.description}")
                
                # MongoDB'ye kaydet
                saved_id = anomaly_repository.save_anomaly(anomaly)
                if saved_id:
                    logger.info(f"Anomali MongoDB'ye kaydedildi: {saved_id}")
                    
                    # SSE üzerinden bildir
                    anomaly_dict = {
                        "id": saved_id,
                        "source": anomaly.source,
                        "timestamp": anomaly.timestamp,
                        "anomaly_type": anomaly.anomaly_type,
                        "pollutant": anomaly.pollutant,
                        "current_value": anomaly.current_value,
                        "average_value": anomaly.average_value,
                        "increase_ratio": anomaly.increase_ratio,
                        "geohash": anomaly.geohash,
                        "latitude": anomaly.latitude,
                        "longitude": anomaly.longitude,
                        "description": anomaly.description
                    }
                    
                    # Async broadcast'i thread-safe şekilde çağır
                    asyncio.run_coroutine_threadsafe(
                        sse_controller.broadcast_anomaly(anomaly_dict),
                        sse_controller_loop
                    )
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

def start_sse_server():
    """SSE sunucusunu başlat"""
    global sse_controller_loop
    sse_controller_loop = asyncio.new_event_loop()
    asyncio.set_event_loop(sse_controller_loop)
    sse_controller_loop.run_until_complete(sse_controller.start_server())

if __name__ == "__main__":
    logger.info("Anomali Tespiti Servisi başlatılıyor...")
    
    # SSE sunucusunu ayrı bir thread'de başlat
    sse_thread = threading.Thread(target=start_sse_server)
    sse_thread.daemon = True
    sse_thread.start()
    
    # Ana thread'de RabbitMQ consumer'ı çalıştır
    start_rabbitmq_consumer()
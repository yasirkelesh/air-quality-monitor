import os
import threading
import uvicorn
from loguru import logger

from infrastructure.config import API_HOST, API_PORT

def start_api_server():
    """API sunucusunu başlat"""
    logger.info("API sunucusu başlatılıyor...")
    from presentation.api.app import app
    uvicorn.run(app, host=API_HOST, port=API_PORT)

def start_rabbitmq_consumer():
    """RabbitMQ consumer'ı başlat"""
    logger.info("RabbitMQ consumer başlatılıyor...")
    # İlerleyen adımlarda implementasyonu yapılacak
    pass

if __name__ == "__main__":
    logger.info("Veri İşleme Servisi başlatılıyor...")
    
    # API'yi ayrı bir thread'de başlat
    api_thread = threading.Thread(target=start_api_server)
    api_thread.daemon = True
    api_thread.start()
    
    # RabbitMQ consumer'ı ana thread'de çalıştır
    start_rabbitmq_consumer()
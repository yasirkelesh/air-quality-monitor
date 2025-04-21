import os
from dotenv import load_dotenv

# .env dosyasını yükle
load_dotenv()

# RabbitMQ Konfigürasyonu
RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "rabbitmq")
RABBITMQ_PORT = int(os.getenv("RABBITMQ_PORT", "5672"))
RABBITMQ_USER = os.getenv("RABBITMQ_USER", "admin")
RABBITMQ_PASS = os.getenv("RABBITMQ_PASS", "password123")
RABBITMQ_ANOMALY_QUEUE = os.getenv("RABBITMQ_ANOMALY_QUEUE", "anomaly-data")
RABBITMQ_PROCESSED_QUEUE = os.getenv("RABBITMQ_PROCESSED_QUEUE", "processed-data")

MONGODB_URI = os.getenv("MONGODB_URI", "mongodb://admin:password@mongodb:27017/data-collector?authSource=admin")
MONGODB_DB = os.getenv("MONGODB_DB", "anomaly_db")
MONGODB_COLLECTION = os.getenv("MONGODB_COLLECTION", "anomalies")

# Anomali yaşam süresi (saat)
ANOMALY_TTL_HOURS = int(os.getenv("ANOMALY_TTL_HOURS", "1"))


# API Konfigürasyonu
API_HOST = os.getenv("API_HOST", "0.0.0.0")
API_PORT = int(os.getenv("API_PORT", "6000"))
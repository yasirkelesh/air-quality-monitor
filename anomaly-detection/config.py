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


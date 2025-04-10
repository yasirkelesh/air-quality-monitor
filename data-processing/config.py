import os
from dotenv import load_dotenv

# .env dosyasını yükle
load_dotenv()

# RabbitMQ Konfigürasyonu
RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "rabbitmq")
RABBITMQ_PORT = int(os.getenv("RABBITMQ_PORT", "5672"))
RABBITMQ_USER = os.getenv("RABBITMQ_USER", "admin")
RABBITMQ_PASS = os.getenv("RABBITMQ_PASS", "password123")
RABBITMQ_RAW_QUEUE = os.getenv("RABBITMQ_RAW_QUEUE", "raw-data")
RABBITMQ_PROCESSED_QUEUE = os.getenv("RABBITMQ_PROCESSED_QUEUE", "processed-data")

# InfluxDB Konfigürasyonu
INFLUXDB_HOST = os.getenv("INFLUXDB_HOST", "localhost")
INFLUXDB_PORT = int(os.getenv("INFLUXDB_PORT", "8086"))
INFLUXDB_USER = os.getenv("INFLUXDB_USER", "")
INFLUXDB_PASS = os.getenv("INFLUXDB_PASS", "")
INFLUXDB_DB = os.getenv("INFLUXDB_DB", "air_quality_db")

# Geocoding Konfigürasyonu
GEOHASH_PRECISION = int(os.getenv("GEOHASH_PRECISION", "5"))

# API Konfigürasyonu
API_HOST = os.getenv("API_HOST", "0.0.0.0")
API_PORT = int(os.getenv("API_PORT", "9000"))
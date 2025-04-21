from pymongo import MongoClient, GEOSPHERE
from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional
from loguru import logger

from config import MONGODB_URI, MONGODB_DB, MONGODB_COLLECTION, ANOMALY_TTL_HOURS
from domain.models.anomaly import Anomaly

class AnomalyRepository:
    """MongoDB'de anomali kayıtlarını yöneten repository"""
    
    def __init__(self):
        """Repository başlat ve bağlantıyı kur"""
        self.client = None
        self.db = None
        self.collection = None
        self.connect()
    
    def connect(self) -> bool:
        """MongoDB'ye bağlan ve indeksleri oluştur"""
        try:
            self.client = MongoClient(MONGODB_URI)
            self.db = self.client[MONGODB_DB]
            self.collection = self.db[MONGODB_COLLECTION]
            
            # Geospatial indeks (harita sorguları için)
            self.collection.create_index([("location", "2dsphere")])
            
            # TTL indeksi (otomatik silme için)
            self.collection.create_index("expiry_time", expireAfterSeconds=0)
            
            # Benzersiz anomali indeksi (aynı anomalinin tekrar kaydedilmesini önler)
            self.collection.create_index(
                [("source", 1), ("pollutant", 1), ("timestamp", 1), ("anomaly_type", 1)],
                unique=True
            )
            
            logger.info(f"MongoDB bağlantısı başarılı: {MONGODB_URI}")
            return True
        
        except Exception as e:
            logger.error(f"MongoDB bağlantı hatası: {str(e)}")
            return False
    
    def save_anomaly(self, anomaly: Anomaly) -> Optional[str]:
        """
        Anomaliyi MongoDB'ye kaydet
        
        Args:
            anomaly: Anomaly nesnesi
            
        Returns:
            Kaydedilen anomalinin ID'si veya None
        """
        try:
            # Anomaliyi dictionary'e dönüştür
            anomaly_dict = {
                "source": anomaly.source,
                "timestamp": anomaly.timestamp,
                "anomaly_type": anomaly.anomaly_type,
                "pollutant": anomaly.pollutant,
                "current_value": anomaly.current_value,
                "average_value": anomaly.average_value,
                "increase_ratio": anomaly.increase_ratio,
                "geohash": anomaly.geohash,
                "geohash_prefix": anomaly.geohash_prefix,
                "location": {
                    "type": "Point",
                    "coordinates": [anomaly.longitude, anomaly.latitude]  # MongoDB GeoJSON formatı
                },
                "country": anomaly.country,
                "city": anomaly.city,
                "district": anomaly.district,
                "description": anomaly.description,
                "detected_at": anomaly.detected_at,
                "expiry_time": datetime.now() + timedelta(hours=ANOMALY_TTL_HOURS)
            }
            
            # MongoDB'ye kaydet
            result = self.collection.insert_one(anomaly_dict)
            logger.info(f"Anomali kaydedildi: {result.inserted_id}")
            return str(result.inserted_id)
        
        except Exception as e:
            if "duplicate key error" in str(e):
                logger.debug(f"Aynı anomali zaten var: {anomaly.source}-{anomaly.pollutant}")
            else:
                logger.error(f"Anomali kaydetme hatası: {str(e)}")
            return None
    
    def get_active_anomalies(self) -> List[Dict[str, Any]]:
        """Aktif (süresi dolmamış) anomalileri getir"""
        try:
            current_time = datetime.now()
            query = {"expiry_time": {"$gt": current_time}}
            
            anomalies = list(self.collection.find(query))
            
            # ObjectId'leri string'e çevir
            for anomaly in anomalies:
                if "_id" in anomaly:
                    anomaly["_id"] = str(anomaly["_id"])
            
            return anomalies
        
        except Exception as e:
            logger.error(f"Aktif anomalileri getirme hatası: {str(e)}")
            return []
    
    def get_anomalies_by_geohash(self, geohash_prefix: str) -> List[Dict[str, Any]]:
        """Belirli bir bölgedeki anomalileri getir"""
        try:
            current_time = datetime.now()
            query = {
                "expiry_time": {"$gt": current_time},
                "geohash_prefix": geohash_prefix
            }
            
            anomalies = list(self.collection.find(query))
            
            for anomaly in anomalies:
                if "_id" in anomaly:
                    anomaly["_id"] = str(anomaly["_id"])
            
            return anomalies
        
        except Exception as e:
            logger.error(f"Geohash bazlı anomali getirme hatası: {str(e)}")
            return []
    
    def close(self):
        """MongoDB bağlantısını kapat"""
        if self.client:
            self.client.close()
            logger.info("MongoDB bağlantısı kapatıldı")
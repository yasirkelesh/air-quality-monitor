from influxdb import InfluxDBClient
from loguru import logger
from typing import Dict, Any, List, Optional
from datetime import datetime

from infrastructure.config import (
    INFLUXDB_HOST, INFLUXDB_PORT, INFLUXDB_USER,
    INFLUXDB_PASS, INFLUXDB_DB
)

class InfluxDBRepository:
    """InfluxDB'ye veri yazma ve okuma işlemlerini yöneten repository"""
    
    def __init__(self):
        """InfluxDB repository'sini başlat"""
        self.client = None
        self.connect()
    
    def connect(self) -> bool:
        """InfluxDB'ye bağlan"""
        try:
            # InfluxDB'ye bağlan
            self.client = InfluxDBClient(
                host=INFLUXDB_HOST,
                port=INFLUXDB_PORT,
                username=INFLUXDB_USER,
                password=INFLUXDB_PASS,
                database=INFLUXDB_DB
            )
            
            # Veritabanını kontrol et, yoksa oluştur
            databases = self.client.get_list_database()
            if {'name': INFLUXDB_DB} not in databases:
                self.client.create_database(INFLUXDB_DB)
            
            # Veritabanını kullan
            self.client.switch_database(INFLUXDB_DB)
            
            logger.info(f"InfluxDB bağlantısı başarılı: {INFLUXDB_HOST}:{INFLUXDB_PORT}/{INFLUXDB_DB}")
            return True
            
        except Exception as e:
            logger.error(f"InfluxDB bağlantı hatası: {str(e)}")
            return False
    
    def save_processed_data(self, data: Dict[str, Any]) -> bool:
        """
        İşlenmiş veriyi InfluxDB'ye kaydet
        
        Args:
            data: İşlenmiş veri sözlüğü
            
        Returns:
            Başarılı oldu mu?
        """
        try:
            # InfluxDB için point oluştur
            point = {
                "measurement": "air_quality",
                "tags": {
                    "source": data.get("source", "unknown"),
                    "geohash": data.get("geohash", ""),
                    "country": data.get("country", ""),
                    "city": data.get("city", ""),
                    "district": data.get("district", "")
                },
                "time": data.get("timestamp"),
                "fields": {
                    "latitude": float(data.get("latitude", 0)),
                    "longitude": float(data.get("longitude", 0)),
                    "pm25": float(data.get("pm25", 0)),
                    "pm10": float(data.get("pm10", 0)),
                    "no2": float(data.get("no2", 0)),
                    "so2": float(data.get("so2", 0)),
                    "o3": float(data.get("o3", 0))
                }
            }
            
            # InfluxDB'ye yaz
            self.client.write_points([point])
            logger.info(f"Veri InfluxDB'ye kaydedildi: {data.get('source', 'unknown')}")
            return True
            
        except Exception as e:
            logger.error(f"InfluxDB yazma hatası: {str(e)}")
            
            # Bağlantı kopmuşsa yeniden bağlanmayı dene
            if not self.client:
                logger.info("Bağlantı yok, yeniden bağlanmayı deniyorum...")
                self.connect()
                
            return False
    
    def get_regional_average(self, geohash: str, hours: int = 24) -> Optional[Dict[str, Any]]:
        """
        Belirli bir geohash bölgesi için ortalama değerleri hesaplar
        
        Args:
            geohash: Sorgulanacak geohash
            hours: Son kaç saatlik veri (varsayılan: 24)
            
        Returns:
            Bölgesel ortalama değerler
        """
        try:
            # Zaman aralığını hesapla
            time_clause = f"time > now() - {hours}h"
            
            # Ortalama değerleri sorgula
            query = f"""
                SELECT 
                    MEAN(pm25) as pm25_avg, 
                    MEAN(pm10) as pm10_avg,
                    MEAN(no2) as no2_avg,
                    MEAN(so2) as so2_avg,
                    MEAN(o3) as o3_avg,
                    COUNT(pm25) as reading_count
                FROM air_quality 
                WHERE geohash = '{geohash}' AND {time_clause}
            """
            
            result = self.client.query(query)
            points = list(result.get_points())
            
            if not points or len(points) == 0:
                logger.warning(f"Geohash için veri bulunamadı: {geohash}")
                return None
            
            # Bölgesel bilgileri sorgula (en son kayıttaki değerleri al)
            location_query = f"""
                SELECT country, city, district
                FROM air_quality
                WHERE geohash = '{geohash}'
                ORDER BY time DESC
                LIMIT 1
            """
            
            location_result = self.client.query(location_query)
            location_points = list(location_result.get_points())
            
            # Sonuçları birleştir
            regional_data = {
                "geohash": geohash,
                "start_time": f"now() - {hours}h",
                "end_time": "now()",
                "reading_count": int(points[0].get("reading_count", 0)),
                "pm25_avg": round(points[0].get("pm25_avg", 0), 2),
                "pm10_avg": round(points[0].get("pm10_avg", 0), 2),
                "no2_avg": round(points[0].get("no2_avg", 0), 2),
                "so2_avg": round(points[0].get("so2_avg", 0), 2),
                "o3_avg": round(points[0].get("o3_avg", 0), 2)
            }
            
            # Eğer konum bilgisi varsa ekle
            if location_points and len(location_points) > 0:
                regional_data["country"] = location_points[0].get("country", "")
                regional_data["city"] = location_points[0].get("city", "")
                regional_data["district"] = location_points[0].get("district", "")
            
            return regional_data
            
        except Exception as e:
            logger.error(f"InfluxDB sorgulama hatası: {str(e)}")
            return None
    
    def get_all_geohashes(self) -> List[str]:
        """
        Sistemdeki tüm geohash'leri listeler
        
        Returns:
            Geohash listesi
        """
        try:
            query = "SHOW TAG VALUES FROM air_quality WITH KEY = geohash"
            result = self.client.query(query)
            
            geohashes = []
            for point in result.get_points():
                if "value" in point:
                    geohashes.append(point["value"])
            
            return geohashes
            
        except Exception as e:
            logger.error(f"Geohash listesi alınamadı: {str(e)}")
            return []
    
    def close(self):
        """Bağlantıyı kapat"""
        if self.client:
            self.client.close()
            logger.info("InfluxDB bağlantısı kapatıldı")
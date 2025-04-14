import influxdb_client
from influxdb_client.client.write_api import SYNCHRONOUS
from loguru import logger
from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta

from config import (
    INFLUXDB_HOST,
    INFLUXDB_PORT,
    INFLUXDB_ORG,
    INFLUXDB_USERNAME,
    INFLUXDB_PASSWORD,
    INFLUXDB_BUCKET,
    INFLUXDB_TOKEN,
)

class InfluxDBRepository:
    """InfluxDB'ye veri yazma ve okuma işlemlerini yöneten repository"""
    
    def __init__(self):
        """InfluxDB repository'sini başlat"""
        self.client = None
        self.write_api = None
        self.query_api = None
        self.connect()
    
    def connect(self) -> bool:
        """InfluxDB'ye bağlan"""
        try:
            # InfluxDB'ye bağlan
            self.client = influxdb_client.InfluxDBClient(
                url=f"http://{INFLUXDB_HOST}:{INFLUXDB_PORT}",
                token=INFLUXDB_TOKEN,
                org=INFLUXDB_ORG,
                username=INFLUXDB_USERNAME,
                password=INFLUXDB_PASSWORD
            )
            
            # API'leri oluştur
            self.write_api = self.client.write_api(write_options=SYNCHRONOUS)
            self.query_api = self.client.query_api()
            
            # Bağlantıyı test et
            health = self.client.health()
            if health.status == "pass":
                logger.info(f"InfluxDB bağlantısı başarılı: {INFLUXDB_HOST}:{INFLUXDB_PORT}")
                return True
            else:
                logger.warning(f"InfluxDB sağlık kontrolü: {health.status}")
                return False
            
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
            # InfluxDB v2 için Flux veri noktası oluştur
            from influxdb_client import Point
            
            point = Point("air_quality") \
                .tag("source", data.get("source", "unknown")) \
                .tag("geohash", data.get("geohash", "")) \
                .tag("country", data.get("country", "")) \
                .tag("city", data.get("city", "")) \
                .tag("district", data.get("district", "")) \
                .field("latitude", float(data.get("latitude", 0))) \
                .field("longitude", float(data.get("longitude", 0))) \
                .field("pm25", float(data.get("pm25", 0))) \
                .field("pm10", float(data.get("pm10", 0))) \
                .field("no2", float(data.get("no2", 0))) \
                .field("so2", float(data.get("so2", 0))) \
                .field("o3", float(data.get("o3", 0)))
            
            # Timestamp değerini ayarla
            timestamp = data.get("timestamp")
            if timestamp:
                point.time(timestamp)
            
            # Veriyi yaz
            self.write_api.write(bucket=INFLUXDB_BUCKET, record=point)
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
            # Zaman aralığını hesapla (şu andan geriye doğru X saat)
            start_time = f"-{hours}h"
            
            # Flux sorgusu oluştur (InfluxDB v2 için Flux dili kullanılır)
            flux_query = f'''
                from(bucket: "{INFLUXDB_BUCKET}")
                    |> range(start: {start_time})
                    |> filter(fn: (r) => r._measurement == "air_quality")
                    |> filter(fn: (r) => r.geohash == "{geohash}")
                    |> filter(fn: (r) => r._field == "pm25" or r._field == "pm10" or 
                                         r._field == "no2" or r._field == "so2" or 
                                         r._field == "o3")
                    |> mean()
                    |> group(columns: ["_field"])
            '''
            
            # Sorguyu çalıştır
            tables = self.query_api.query(query=flux_query, org=INFLUXDB_ORG)
            
            # Sonuç tablosunu işle
            if not tables or len(tables) == 0:
                logger.warning(f"Geohash için veri bulunamadı: {geohash}")
                return None
            
            # Değerleri topla
            readings = {}
            for table in tables:
                for record in table.records:
                    field = record.get_field()
                    value = record.get_value()
                    if field and value is not None:
                        readings[f"{field}_avg"] = round(float(value), 2)
            
            # Kayıt sayısını al
            count_query = f'''
                from(bucket: "{INFLUXDB_BUCKET}")
                    |> range(start: {start_time})
                    |> filter(fn: (r) => r._measurement == "air_quality")
                    |> filter(fn: (r) => r.geohash == "{geohash}")
                    |> filter(fn: (r) => r._field == "pm25")
                    |> count()
            '''
            
            count_tables = self.query_api.query(query=count_query, org=INFLUXDB_ORG)
            reading_count = 0
            
            if count_tables and len(count_tables) > 0:
                for table in count_tables:
                    for record in table.records:
                        reading_count = int(record.get_value() or 0)
            
            # Konum bilgilerini getir
            location_query = f'''
                from(bucket: "{INFLUXDB_BUCKET}")
                    |> range(start: {start_time})
                    |> filter(fn: (r) => r._measurement == "air_quality")
                    |> filter(fn: (r) => r.geohash == "{geohash}")
                    |> last()
                    |> keep(columns: ["country", "city", "district"])
                    |> limit(n: 1)
            '''
            
            location_tables = self.query_api.query(query=location_query, org=INFLUXDB_ORG)
            location = {}
            
            if location_tables and len(location_tables) > 0:
                for table in location_tables:
                    for record in table.records:
                        location["country"] = record.values.get("country", "")
                        location["city"] = record.values.get("city", "")
                        location["district"] = record.values.get("district", "")
            
            # Sonuçları birleştir
            regional_data = {
                "geohash": geohash,
                "start_time": datetime.now() - timedelta(hours=hours),
                "end_time": datetime.now(),
                "reading_count": reading_count,
                "pm25_avg": readings.get("pm25_avg", 0),
                "pm10_avg": readings.get("pm10_avg", 0),
                "no2_avg": readings.get("no2_avg", 0),
                "so2_avg": readings.get("so2_avg", 0),
                "o3_avg": readings.get("o3_avg", 0)
            }
            
            # Konum bilgilerini ekle
            regional_data.update(location)
            
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
            # InfluxDB v2 için Flux sorgusu
            flux_query = f'''
                import "influxdata/influxdb/schema"
                
                schema.tagValues(
                    bucket: "{INFLUXDB_BUCKET}",
                    tag: "geohash",
                    predicate: (r) => r._measurement == "air_quality"
                )
            '''
            
            tables = self.query_api.query(query=flux_query, org=INFLUXDB_ORG)
            
            geohashes = []
            for table in tables:
                for record in table.records:
                    value = record.get_value()
                    if value:
                        geohashes.append(value)
            
            return geohashes
            
        except Exception as e:
            logger.error(f"Geohash listesi alınamadı: {str(e)}")
            return []
    
    def close(self):
        """Bağlantıyı kapat"""
        if self.client:
            self.client.close()
            logger.info("InfluxDB bağlantısı kapatıldı")
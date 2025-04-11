from datetime import datetime
from loguru import logger
from typing import Dict, Any

from domain.models.raw_data import RawData
from domain.models.processed_data import ProcessedData
from business.geocoding_service import GeocodingService

class DataProcessor:
    """Ham veriyi işleyerek zenginleştiren servis"""
    
    def __init__(self):
        """Veri işleme servisini başlat"""
        self.geocoding_service = GeocodingService()
    
    def process(self, raw_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Ham veriyi işler ve zenginleştirir
        
        Args:
            raw_data: Ham veri sözlüğü
            
        Returns:
            İşlenmiş veri sözlüğü
        """
        try:
            # Ham veriyi doğrula
            validated_data = self._validate_raw_data(raw_data)
            
            # Geohash oluştur
            geohash = self.geocoding_service.get_geohash(
                validated_data.latitude, 
                validated_data.longitude
            )
            
            # Adres bilgileri getir
            address = self.geocoding_service.get_address(
                validated_data.latitude, 
                validated_data.longitude
            )
            
            # İşlenmiş veriyi oluştur
            processed_data = ProcessedData(
                # Orijinal sensör verileri
                latitude=validated_data.latitude,
                longitude=validated_data.longitude,
                timestamp=validated_data.timestamp,
                pm25=validated_data.pm25,
                pm10=validated_data.pm10,
                no2=validated_data.no2,
                so2=validated_data.so2,
                o3=validated_data.o3,
                source=validated_data.source,
                
                # İşleme sonucu eklenen alanlar
                geohash=geohash,
                country=address.get("country"),
                city=address.get("city"),
                district=address.get("district"),
                processed_at=datetime.now()
            )
            
            return processed_data.dict()
            
        except Exception as e:
            logger.error(f"Veri işleme hatası: {str(e)}")
            raise
    
    def _validate_raw_data(self, raw_data: Dict[str, Any]) -> RawData:
        """
        Ham veriyi doğrular ve RawData modeline dönüştürür
        
        Args:
            raw_data: Ham veri sözlüğü
            
        Returns:
            Doğrulanmış RawData nesnesi
        """
        try:
            # MongoDB'den gelen `_id` alanını `id` olarak yeniden adlandır
            if '_id' in raw_data and 'id' not in raw_data:
                raw_data['id'] = str(raw_data.pop('_id'))
            
            # timestamp alanını kontrol et
            if 'timestamp' in raw_data and isinstance(raw_data['timestamp'], dict):
                # MongoDB tarih formatını kontrol et
                if '$date' in raw_data['timestamp']:
                    raw_data['timestamp'] = raw_data['timestamp']['$date']
            
            # RawData modeli ile doğrula
            return RawData(**raw_data)
            
        except Exception as e:
            logger.error(f"Veri doğrulama hatası: {str(e)}")
            raise
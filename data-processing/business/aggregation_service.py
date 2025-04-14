from loguru import logger
from typing import Dict, Any, List, Optional

from data_access.influxdb_repository import InfluxDBRepository

class AggregationService:
    """Bölgesel ortalama değerleri hesaplayan servis"""
    
    def __init__(self):
        """Agregasyon servisini başlat"""
        self.repository = InfluxDBRepository()
    
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
            return self.repository.get_regional_average(geohash, hours)
        except Exception as e:
            logger.error(f"Bölgesel ortalama hesaplama hatası: {str(e)}")
            return None
    
    def get_all_regional_averages(self, hours: int = 24) -> List[Dict[str, Any]]:
        """
        Tüm geohash bölgeleri için ortalama değerleri hesaplar
        
        Args:
            hours: Son kaç saatlik veri (varsayılan: 24)
            
        Returns:
            Tüm bölgelerin ortalama değerleri
        """
        try:
            # Tüm geohash'leri al
            geohashes = self.repository.get_all_geohashes()
            
            # Her geohash için ortalama değerleri hesapla
            results = []
            for geohash in geohashes:
                avg_data = self.repository.get_regional_average(geohash, hours)
                if avg_data:
                    results.append(avg_data)
            
            return results
        except Exception as e:
            logger.error(f"Tüm bölgesel ortalamalar hesaplanamadı: {str(e)}")
            return []
# business/anomaly_detector.py
from typing import List, Dict, Any
from loguru import logger
from domain.models.processed_data import ProcessedData
from domain.models.anomaly import Anomaly, AnomalyType

class AnomalyDetector:
    """Basit anomali tespiti yapan sınıf"""
    
    def __init__(self):
        # Anomali eşik değerleri
        self.time_series_threshold = 0.5   # %50 artış
        self.spatial_threshold = 1.0       # %100 artış (2x)
    
    def detect(self, data: ProcessedData) -> List[Anomaly]:
        """
        Anomali tespiti yap
        
        Args:
            data: İşlenmiş sensör verisi
            
        Returns:
            Tespit edilen anomaliler listesi
        """
        anomalies = []
        
        # Regional average yoksa anomali tespiti yapılamaz
        if not data.regional_average:
            logger.warning(f"Regional average bulunamadı: {data.source}")
            return anomalies
        
        # Her kirletici için anomali kontrolü
        for pollutant in ["pm25", "pm10", "no2", "so2", "o3"]:
            current_value = getattr(data, pollutant, None)
            avg_value = data.regional_average.get(f"{pollutant}_avg", None)
            
            if current_value and avg_value and avg_value > 0:
                # Artış oranını hesapla
                increase_ratio = (current_value - avg_value) / avg_value
                
                # Zaman serisi anomalisi kontrolü (%50+ artış)
                if increase_ratio >= self.time_series_threshold:
                    anomaly = Anomaly(
                        geohash=data.geohash,
                        source=data.source,
                        timestamp=data.timestamp,
                        anomaly_type=AnomalyType.TIME_SERIES,
                        pollutant=pollutant,
                        current_value=current_value,
                        average_value=avg_value,
                        increase_ratio=increase_ratio,
                        latitude=data.latitude,
                        longitude=data.longitude,
                        country=data.country,
                        city=data.city,
                        district=data.district,
                        description=f"{pollutant.upper()} %{increase_ratio*100:.1f} artış gösterdi"
                    )
                    anomalies.append(anomaly)
                    logger.warning(f"[TIME_SERIES] {data.source}: {pollutant} %{increase_ratio*100:.1f} artış")
                
                # Mekansal anomali kontrolü (%100+ artış)
                if increase_ratio >= self.spatial_threshold:
                    anomaly = Anomaly(
                        geohash=data.geohash,
                        source=data.source,
                        timestamp=data.timestamp,
                        anomaly_type=AnomalyType.SPATIAL,
                        pollutant=pollutant,
                        current_value=current_value,
                        average_value=avg_value,
                        increase_ratio=increase_ratio,
                        latitude=data.latitude,
                        longitude=data.longitude,
                        country=data.country,
                        city=data.city,
                        district=data.district,
                        description=f"{pollutant.upper()} bölge ortalamasının {increase_ratio+1:.1f} katı"
                    )
                    anomalies.append(anomaly)
                    logger.warning(f"[SPATIAL] {data.source}: {pollutant} {increase_ratio+1:.1f}x ortalama")
        
        return anomalies
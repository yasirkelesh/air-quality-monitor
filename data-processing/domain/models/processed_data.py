from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime

class ProcessedData(BaseModel):
    """İşlenmiş hava kalitesi verisi modeli"""
    # Orijinal sensör verileri
    latitude: float
    longitude: float
    timestamp: datetime
    pm25: Optional[float] = None
    pm10: Optional[float] = None
    no2:  Optional[float] = None
    so2:  Optional[float] = None
    o3:   Optional[float] = None
    source: str
    
    # İşleme sonucu eklenen alanlar
    geohash: str
    country: Optional[str] = None
    city: Optional[str] = None
    district: Optional[str] = None
    processed_at: datetime #isleme ile uretme arasinda fark var mi ?




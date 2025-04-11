from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime

class ProcessedData(BaseModel):
    """İşlenmiş hava kalitesi verisi modeli"""
    # Orijinal sensör verileri
    latitude: float
    longitude: float
    timestamp: datetime
    pm25: float
    pm10: float
    no2: float
    so2: float
    o3: float
    source: str
    
    # İşleme sonucu eklenen alanlar
    geohash: str
    country: Optional[str] = None
    city: Optional[str] = None
    district: Optional[str] = None
    processed_at: datetime #isleme ile uretme arasinda fark var mi ?
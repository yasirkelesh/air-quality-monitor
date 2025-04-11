from pydantic import BaseModel
from typing import Dict, Optional, List
from datetime import datetime

class RegionalAverage(BaseModel):
    """Bölgesel ortalama değerleri"""
    geohash: str
    country: Optional[str] = None
    city: Optional[str] = None
    district: Optional[str] = None
    start_time: datetime
    end_time: datetime
    reading_count: int
    pm25_avg: float
    pm10_avg: float
    no2_avg: float
    so2_avg: float
    o3_avg: float
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
    pm25_avg: Optional[float] = None
    pm10_avg: Optional[float] = None
    no2_avg:  Optional[float] = None
    so2_avg:  Optional[float] = None
    o3_avg:   Optional[float] = None
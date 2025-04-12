from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime

class RawData(BaseModel):
    """Ham hava kalitesi verisi modeli"""
    latitude: float
    longitude: float
    timestamp: datetime
    pm25: Optional[float] = None
    pm10: Optional[float] = None
    no2: Optional[float] = None
    so2: Optional[float] = None
    o3: Optional[float] = None
    source: str



    {
     'id': '67f97ce4c0f87156f17d68d6', 
     'latitude': 41.0082, 
     'longitude': 28.9784, 
     'timestamp': '2025-04-11T20:34:44.968811094Z', 
     'pm25': 18.5, 
     'pm10': 35.2, 
     'source': 'api'
     }
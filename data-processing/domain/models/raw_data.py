from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime

class RawData(BaseModel):
    """Ham hava kalitesi verisi modeli"""
    latitude: float
    longitude: float
    timestamp: datetime
    pm25: float
    pm10: float
    no2: float
    so2: float
    o3: float
    source: str
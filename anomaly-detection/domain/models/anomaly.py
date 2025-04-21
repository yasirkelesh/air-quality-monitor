# domain/models/anomaly.py
from datetime import datetime
from typing import List, Optional
from enum import Enum

class AnomalyType(str, Enum):
    TIME_SERIES = "TIME_SERIES"    # %50+ artış
    SPATIAL = "SPATIAL"            # %100+ artış (mekansal)

class Anomaly:
    def __init__(
        self,
        geohash: str,
        source: str,
        timestamp: str,
        anomaly_type: AnomalyType,
        pollutant: str,
        current_value: float,
        average_value: float,
        increase_ratio: float,
        latitude: float,
        longitude: float,
        country: Optional[str] = None,
        city: Optional[str] = None,
        district: Optional[str] = None,
        description: str = ""
    ):
        self.geohash = geohash
        self.geohash_prefix = geohash[:3] if len(geohash) >= 3 else geohash
        self.source = source
        self.timestamp = timestamp
        self.anomaly_type = anomaly_type
        self.pollutant = pollutant
        self.current_value = current_value
        self.average_value = average_value
        self.increase_ratio = increase_ratio
        self.latitude = latitude
        self.longitude = longitude
        self.country = country
        self.city = city
        self.district = district
        self.description = description
        self.detected_at = datetime.now().isoformat()
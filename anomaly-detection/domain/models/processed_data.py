# domain/models/processed_data.py
from datetime import datetime
from typing import Optional, Dict, Any

class ProcessedData:
    def __init__(self, data: Dict[str, Any]):
        self.latitude = data.get('latitude')
        self.longitude = data.get('longitude')
        self.timestamp = data.get('timestamp')
        self.pm25 = data.get('pm25')
        self.pm10 = data.get('pm10')
        self.no2 = data.get('no2')
        self.so2 = data.get('so2')
        self.o3 = data.get('o3')
        self.source = data.get('source')
        self.geohash = data.get('geohash')
        self.country = data.get('country')
        self.city = data.get('city')
        self.district = data.get('district')
        self.processed_at = data.get('processed_at')
        self.regional_average = data.get('regional_average')
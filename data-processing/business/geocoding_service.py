import geohash
from geopy.geocoders import Nominatim
from loguru import logger
from typing import Dict, Any, Optional
from config import GEOHASH_PRECISION

class GeocodingService:
    """Koordinatları adres bilgilerine dönüştüren servis"""
    
    def __init__(self):
        """Geocoding servisini başlat"""
        self.geocoder = Nominatim(user_agent="veri_isleme_servisi")
        self.cache = {}  # Basit önbellek bu kisim gelistirilebilir 
    
    def get_geohash(self, latitude: float, longitude: float) -> str:
        """
        Koordinatlar için geohash üretir
        
        Args:
            latitude: Enlem
            longitude: Boylam
            
        Returns:
            Geohash string
        """
        try:
            return geohash.encode(latitude, longitude, precision=GEOHASH_PRECISION)
        except Exception as e:
            logger.error(f"Geohash oluşturma hatası: {str(e)}")
            return ""
    
    def get_address(self, latitude: float, longitude: float) -> Dict[str, Any]:
        """
        Koordinatları adres bilgilerine dönüştürür
        
        Args:
            latitude: Enlem
            longitude: Boylam
            
        Returns:
            Adres bilgileri içeren sözlük
        """
        # Önbellekte var mı kontrol et
        cache_key = f"{latitude:.6f},{longitude:.6f}"
        if cache_key in self.cache:
            return self.cache[cache_key]
        
        try:
            location = self.geocoder.reverse(f"{latitude}, {longitude}", language="en")
            
            if location and location.raw.get("address"):
                address = location.raw["address"]
                
                # Adres bilgilerini çıkar
                result = {
                    "country": address.get("country"),
                    "city": address.get("city") or address.get("state") or address.get("county"),
                    "district": address.get("suburb") or address.get("town") or address.get("village")
                }
                
                # Önbelleğe al
                self.cache[cache_key] = result
                return result
            else:
                logger.warning(f"Adres bulunamadı: {latitude}, {longitude}")
                return {"country": None, "city": None, "district": None}
                
        except Exception as e:
            logger.error(f"Geocoding hatası: {str(e)}")
            return {"country": None, "city": None, "district": None}
// LocationAnalysisPanel.js
import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './LocationAnalysisPanel.css';

// İkon bileşenleri
const CloseIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <line x1="18" y1="6" x2="6" y2="18"></line>
    <line x1="6" y1="6" x2="18" y2="18"></line>
  </svg>
);

const SearchIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <circle cx="11" cy="11" r="8"></circle>
    <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
  </svg>
);

const LocationMarkerIcon = () => (
  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
    <circle cx="12" cy="10" r="3"></circle>
  </svg>
);

const LocationAnalysisPanel = ({ isOpen, onClose, selectedLocation, airQualityData, selectedMetric }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [locationDetails, setLocationDetails] = useState(null);
  const [loading, setLoading] = useState(false);
  const [nearbyPoints, setNearbyPoints] = useState([]);

  // Seçilen lokasyon değiştiğinde veri yükleme
  useEffect(() => {
    if (selectedLocation && isOpen) {
      setLoading(true);
      
      // Konum detaylarını hazırla
      const locationInfo = {
        coordinates: selectedLocation,
        address: "Yükleniyor...",
        airQuality: {}
      };
      
      setLocationDetails(locationInfo);
      
      // Yakındaki hava kalitesi nokta verilerini bul
      findNearbyAirQualityPoints(selectedLocation);
      
      // Ters geocoding ile adres bilgisini al (örnek)
      fetchLocationAddress(selectedLocation)
        .then(address => {
          setLocationDetails(prev => ({
            ...prev,
            address: address
          }));
          setLoading(false);
        })
        .catch(err => {
          console.error("Adres bilgisi alınamadı:", err);
          setLocationDetails(prev => ({
            ...prev,
            address: "Adres bilgisi bulunamadı"
          }));
          setLoading(false);
        });
    }
  }, [selectedLocation, isOpen]);

  // Ters geocoding (örnek fonksiyon - gerçek uygulamada API kullanılacak)
  const fetchLocationAddress = async (coordinates) => {
    // Bu örnek için geciktirme ekliyoruz, gerçek uygulamada bir API kullanılmalı
    return new Promise((resolve) => {
      setTimeout(() => {
        // Örnek adres yanıtı
        const [lng, lat] = coordinates;
        resolve(`${lat.toFixed(5)}°N, ${lng.toFixed(5)}°E yakınlarında`);
      }, 1000);
    });
  };

  // Yakındaki hava kalitesi noktalarını bul
  const findNearbyAirQualityPoints = (coordinates) => {
    if (!airQualityData || airQualityData.length === 0) {
      setNearbyPoints([]);
      return;
    }

    // Kullanıcının tıkladığı konum
    const [userLng, userLat] = coordinates;
    
    // Her veri noktası için mesafeyi hesapla
    const pointsWithDistance = airQualityData.map(point => {
      const [pointLng, pointLat] = decodeGeohash(point.geohash);
      
      // Haversine formülü ile iki nokta arasındaki mesafeyi hesapla (km cinsinden)
      const distance = calculateDistance(userLat, userLng, pointLat, pointLng);
      
      return {
        ...point,
        distance,
        coordinates: [pointLng, pointLat]
      };
    });
    
    // Mesafeye göre sırala ve en yakın 5 noktayı al
    const closest = pointsWithDistance
      .sort((a, b) => a.distance - b.distance)
      .slice(0, 5);
    
    setNearbyPoints(closest);
  };

  // İki nokta arasındaki mesafeyi hesaplama (Haversine formülü)
  const calculateDistance = (lat1, lon1, lat2, lon2) => {
    const R = 6371; // Dünya yarıçapı km
    const dLat = deg2rad(lat2 - lat1);
    const dLon = deg2rad(lon2 - lon1);
    const a = 
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(deg2rad(lat1)) * Math.cos(deg2rad(lat2)) * 
      Math.sin(dLon / 2) * Math.sin(dLon / 2); 
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a)); 
    const distance = R * c; // Mesafe (km)
    return distance;
  };

  const deg2rad = (deg) => {
    return deg * (Math.PI / 180);
  };

  // Geohash çözme (gerçek uygulamada daha doğru bir kütüphane kullanılmalı)
  const decodeGeohash = (geohash) => {
    // Basitleştirilmiş geohash decoder
    if (geohash === 'sxjrb') {
      return [29.91, 40.76]; // Kocaeli için yaklaşık konum
    } else if (geohash === 'y5642') {
      return [92.79, 56.01]; // Krasnoyarsk için yaklaşık konum
    } else if (geohash === 'u4pruydqqvj8') {
      return [13.41, 52.52]; // Berlin için yaklaşık konum
    } else if (geohash === '9q5ctr') {
      return [-118.24, 34.05]; // Los Angeles için yaklaşık konum
    } else if (geohash === 'wtw3sm') {
      return [151.21, -33.87]; // Sydney için yaklaşık konum
    }
    
    // Diğer geohash'ler için basit yaklaşım
    const hash = geohash.split('');
    const lng = (hash.reduce((a, c) => a + c.charCodeAt(0), 0) % 360) - 180;
    const lat = (hash.reduce((a, c) => a + c.charCodeAt(0) * 2, 0) % 180) - 90;
    return [lng, lat];
  };


  // AQI rengini hesapla
  const getAQIColor = (value, metric) => {
    if (metric === 'pm25_avg') {
      if (value < 12) return '#00e400'; // İyi
      if (value < 35.4) return '#ffff00'; // Orta
      if (value < 55.4) return '#ff7e00'; // Hassas gruplar için sağlıksız
      if (value < 150.4) return '#ff0000'; // Sağlıksız
      if (value < 250.4) return '#99004c'; // Çok sağlıksız
      return '#7e0023'; // Tehlikeli
    }
    // Diğer metrikler için benzer renkler...
    return '#999';
  };

  // AQI seviyesi metni
  const getAQILevel = (value, metric) => {
    if (metric === 'pm25_avg') {
      if (value < 12) return 'İyi';
      if (value < 35.4) return 'Orta';
      if (value < 55.4) return 'Hassas Gruplar İçin Sağlıksız';
      if (value < 150.4) return 'Sağlıksız';
      if (value < 250.4) return 'Çok Sağlıksız';
      return 'Tehlikeli';
    }
    // Diğer metrikler için benzer seviyeler...
    return 'Bilinmiyor';
  };

  // Metrik adını formatla
  const formatMetricName = (metric) => {
    switch(metric) {
      case 'pm25_avg': return 'PM2.5';
      case 'pm10_avg': return 'PM10';
      case 'no2_avg': return 'NO₂';
      case 'so2_avg': return 'SO₂';
      case 'o3_avg': return 'O₃';
      default: return metric;
    }
  };

  if (!isOpen) return null;

  return (
    <div className="location-analysis-panel">
      <div className="panel-header">
        <h3>Konum Analizi</h3>
        <button className="close-button" onClick={onClose}>
          <CloseIcon />
        </button>
      </div>

 
      {locationDetails && (
        <div className="location-details">
          <div className="coordinates-display">
            <LocationMarkerIcon />
            <span>{`${locationDetails.coordinates[1].toFixed(6)}°N, ${locationDetails.coordinates[0].toFixed(6)}°E`}</span>
          </div>
          
          <div className="address-display">
            {loading ? "Yükleniyor..." : locationDetails.address}
          </div>

          <div className="section-divider"></div>

          <h4>Hava Kalitesi Analizi</h4>
          
          {nearbyPoints.length > 0 ? (
            <div className="air-quality-analysis">
              <p>Bu konuma en yakın {nearbyPoints.length} ölçüm noktası:</p>
              
              <div className="nearby-points-list">
                {nearbyPoints.map((point, index) => (
                  <div className="nearby-point-item" key={index}>
                    <div className="point-header">
                      <span className="point-name">{point.city}{point.district ? `, ${point.district}` : ''}</span>
                      <span className="point-distance">{point.distance.toFixed(1)} km</span>
                    </div>
                    
                    <div className="point-metrics">
                      <div className="metric-box" 
                        style={{ backgroundColor: getAQIColor(point[selectedMetric], selectedMetric) }}>
                        <div className="metric-value">{point[selectedMetric].toFixed(1)}</div>
                        <div className="metric-name">{formatMetricName(selectedMetric)}</div>
                      </div>
                      
                      <div className="metric-details">
                        <div className="aqi-level">{getAQILevel(point[selectedMetric], selectedMetric)}</div>
                        <div className="reading-info">
                          <span>Okuma: {point.reading_count}</span>
                          <span>•</span>
                          <span>{new Date(point.start_time).toLocaleDateString()}</span>
                        </div>
                      </div>
                    </div>
                    
                    <div className="all-metrics">
                      <div className="metric-row">
                        <span>PM2.5: {point.pm25_avg.toFixed(1)} µg/m³</span>
                        <span>PM10: {point.pm10_avg.toFixed(1)} µg/m³</span>
                      </div>
                      <div className="metric-row">
                        <span>NO₂: {point.no2_avg.toFixed(1)} µg/m³</span>
                        <span>SO₂: {point.so2_avg.toFixed(1)} µg/m³</span>
                        <span>O₃: {point.o3_avg.toFixed(1)} µg/m³</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="no-data-message">
              <p>Bu konumun yakınında hava kalitesi verisi bulunamadı.</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default LocationAnalysisPanel;
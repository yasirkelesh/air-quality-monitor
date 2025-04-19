// App.js içinde yapılacak değişiklikler
import React, { useState, useEffect, useRef, useCallback } from 'react';
import mapboxgl from 'mapbox-gl';
import axios from 'axios';
import Geohash from 'latlon-geohash'; // Geohash kütüphanesi
import 'mapbox-gl/dist/mapbox-gl.css';
import './App.css';

// Mapbox token
mapboxgl.accessToken = 'pk.eyJ1IjoibXVrZWxlcyIsImEiOiJjbTlpOGRiazcwMDF3MmtzZDUzc3VvZ3k1In0.yKyyKGnzLeQU9kreNrw8MA';


function App() {
  const mapContainer = useRef(null);
  const map = useRef(null);
  const [airQualityData, setAirQualityData] = useState([]);
  const [selectedMetric, setSelectedMetric] = useState('pm25_avg');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showHeatmap, setShowHeatmap] = useState(true); // Varsayılan olarak açık
  const [heatmapIntensity, setHeatmapIntensity] = useState(1.5); // Yoğunluk kontrolü

  const [autoUpdate, setAutoUpdate] = useState(false); // Başlangıçta kapalı
  const [updateInterval, setUpdateInterval] = useState(60); // Varsayılan 60 saniye
  const [lastUpdated, setLastUpdated] = useState(null);
  const intervalRef = useRef(null); // setInterval referansını saklamak için


   // Veri çekme fonksiyonu - useCallback ile memoize ediliyor
   const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      // API URL'ini environment değişkeninden al
      const response = await axios.get('http://localhost:5000/regional-averages');
      
      setAirQualityData(response.data);
      setLastUpdated(new Date());
      setLoading(false);
      setError(null);
    } catch (err) {
      setError('Veriler yüklenirken hata oluştu.');
      setLoading(false);
      console.error('Veri çekme hatası:', err);
    }
  }, []);

  // İlk veri yüklemesi
  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Otomatik güncelleme zamanlayıcısı
  useEffect(() => {
    // Önceki interval'ı temizle
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    
    // Eğer autoUpdate açıksa yeni interval oluştur
    if (autoUpdate) {
      intervalRef.current = setInterval(() => {
        fetchData();
      }, updateInterval * 1000);
      
      console.log(`Otomatik güncelleme aktif: ${updateInterval} saniye aralıkla`);
    }
    
    // Component unmount olduğunda veya dependency değiştiğinde interval'ı temizle
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [autoUpdate, updateInterval, fetchData]);
 

  // Metriklere göre renk seçimi
  const getColor = (value, metric) => {
    // AQI skalasına yakın renkler (metriğe göre ayarlanabilir)
    if (metric === 'pm25_avg') {
      if (value < 12) return '#00e400'; // İyi
      if (value < 35.4) return '#ffff00'; // Orta
      if (value < 55.4) return '#ff7e00'; // Hassas gruplar için sağlıksız
      if (value < 150.4) return '#ff0000'; // Sağlıksız
      if (value < 250.4) return '#99004c'; // Çok sağlıksız
      return '#7e0023'; // Tehlikeli
    } else if (metric === 'pm10_avg') {
      if (value < 54) return '#00e400';
      if (value < 154) return '#ffff00';
      if (value < 254) return '#ff7e00';
      if (value < 354) return '#ff0000';
      if (value < 424) return '#99004c';
      return '#7e0023';
    } else if (metric === 'no2_avg' || metric === 'so2_avg' || metric === 'o3_avg') {
      // Diğer kirleticiler için basit bir skala
      if (value < 20) return '#00e400';
      if (value < 40) return '#ffff00';
      if (value < 60) return '#ff7e00';
      if (value < 80) return '#ff0000';
      if (value < 100) return '#99004c';
      return '#7e0023';
    }
    return '#999'; // Varsayılan
  };

  // Haritayı başlatma
  useEffect(() => {
    if (map.current) return; // Haritayı tekrar başlatmayı önle
    
    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/standard-satellite', // Koyu tema harita
      center: [30.7133, 40.7667], // kullanici konumu
      zoom: 3 // Global görünüm için zoom seviyesi
    });

    map.current.addControl(new mapboxgl.NavigationControl(), 'top-right');
    
    // Harita yüklendiğinde
    map.current.on('load', () => {
      // Özel ısı haritası için ekstra harita kaynaklarını ekle
      // Sembolik bir raster kaynak oluştur (arka plan layer için)
      map.current.addSource('empty-source', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: []
        }
      });

      // Yoğunluk haritası için arka plan layer'ı ekle
      map.current.addLayer({
        id: 'intensity-background',
        type: 'background',
        layout: {
          visibility: 'none' // Başlangıçta gizli
        },
        paint: {
          'background-color': 'rgba(0, 0, 0, 0)'
        }
      });

      // GeoJSON veri kaynağı
      map.current.addSource('air-quality-data', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: []
        }
      });

      // Görüntüdeki gibi ısı haritası (kırmızı/turuncu gradient) 
      map.current.addLayer({
        id: 'air-quality-heat',
        type: 'heatmap',
        source: 'air-quality-data',
        layout: {
          visibility: 'visible' // Başlangıçta görünür
        },
        paint: {
          // Isı haritası ağırlık özelliği
          'heatmap-weight': [
            'interpolate',
            ['linear'],
            ['get', 'value'],
            0, 0,
            100, 1
          ],
          // Isı haritası yoğunluğu (yüksek değer daha yoğun görünüm)
          'heatmap-intensity': [
            'interpolate',
            ['linear'],
            ['zoom'],
            0, 1,
            9, 5
          ],
          // Görseldeki gibi, yoğun kırmızı/turuncu/sarı renk dağılımı
          'heatmap-color': [
            'interpolate',
            ['linear'],
            ['heatmap-density'],
            0, 'rgba(236, 222, 139, 0)',
            0.1, 'rgba(236, 222, 139, 0.5)', // açık sarı
            0.3, 'rgba(255, 211, 0, 0.7)',   // koyu sarı
            0.5, 'rgba(255, 140, 0, 0.8)',   // turuncu
            0.7, 'rgba(220, 38, 38, 0.8)',   // kırmızı
            0.9, 'rgba(140, 20, 20, 0.9)',   // koyu kırmızı
            1, 'rgba(80, 10, 10, 0.9)'       // çok koyu kırmızı
          ],
          // Isı haritası yarıçapı - görseldeki gibi geniş bir etki alanı
          'heatmap-radius': [
            'interpolate',
            ['linear'],
            ['zoom'],
            0, 15,
            4, 25,
            9, 50
          ],
          // Isı haritası opaklığı - daha yüksek opaklık
          'heatmap-opacity': 0.9
        }
      });

      // Noktalar katmanı
      map.current.addLayer({
        id: 'air-quality-point',
        type: 'circle',
        source: 'air-quality-data',
        paint: {
          'circle-radius': [
            'interpolate',
            ['linear'],
            ['zoom'],
            3, 6,
            9, 12
          ],
          'circle-color': ['get', 'color'],
          'circle-stroke-width': 1,
          'circle-stroke-color': '#fff'
        }
      });

      // Popup
      const popup = new mapboxgl.Popup({
        closeButton: false,
        closeOnClick: false
      });

      map.current.on('mouseenter', 'air-quality-point', (e) => {
        map.current.getCanvas().style.cursor = 'pointer';
        
        const coordinates = e.features[0].geometry.coordinates.slice();
        const properties = e.features[0].properties;
        
        const html = `
          <h3>${properties.city}${properties.district ? ', ' + properties.district : ''}</h3>
          <p>${properties.country}</p>
          <p><strong>PM2.5:</strong> ${properties.pm25_avg} µg/m³</p>
          <p><strong>PM10:</strong> ${properties.pm10_avg} µg/m³</p>
          <p><strong>NO2:</strong> ${properties.no2_avg} µg/m³</p>
          <p><strong>SO2:</strong> ${properties.so2_avg} µg/m³</p>
          <p><strong>O3:</strong> ${properties.o3_avg} µg/m³</p>
          <p><strong>Okuma sayısı:</strong> ${properties.reading_count}</p>
          <p><strong>Zaman aralığı:</strong><br>${new Date(properties.start_time).toLocaleDateString()} - ${new Date(properties.end_time).toLocaleDateString()}</p>
        `;

        popup.setLngLat(coordinates)
          .setHTML(html)
          .addTo(map.current);
      });

      map.current.on('mouseleave', 'air-quality-point', () => {
        map.current.getCanvas().style.cursor = '';
        popup.remove();
      });
    });
  }, []);

  // Verileri haritaya yükleme
  useEffect(() => {
    if (!map.current || !map.current.isStyleLoaded() || airQualityData.length === 0) return;

    // Daha fazla veri noktası oluştur (gerçek verileri artır)
    const expandedData = expandAirQualityData(airQualityData);
    
    // GeoJSON veri yapısı oluştur
    const geojsonData = {
      type: 'FeatureCollection',
      features: expandedData.map(location => {
        // Geohash'ten koordinat dönüşümü
        const [lng, lat] = decodeGeohash(location.geohash);
        
        // Metrik değerini normalize et (0-100 arası)
        let normalizedValue;
        
        if (selectedMetric === 'pm25_avg') {
          normalizedValue = Math.min(100, (location[selectedMetric] / 250) * 100);
        } else if (selectedMetric === 'pm10_avg') {
          normalizedValue = Math.min(100, (location[selectedMetric] / 400) * 100);
        } else {
          normalizedValue = Math.min(100, (location[selectedMetric] / 100) * 100);
        }
        
        return {
          type: 'Feature',
          geometry: {
            type: 'Point',
            coordinates: [lng, lat]
          },
          properties: {
            ...location,
            color: getColor(location[selectedMetric], selectedMetric),
            value: normalizedValue * heatmapIntensity // Isı haritası yoğunluk ayarı
          }
        };
      })
    };

    // Veri kaynağını güncelle
    if (map.current.getSource('air-quality-data')) {
      map.current.getSource('air-quality-data').setData(geojsonData);
    }
    
  }, [airQualityData, selectedMetric, heatmapIntensity]);

  // Isı haritasını göster/gizle
  useEffect(() => {
    if (!map.current || !map.current.isStyleLoaded()) return;
    
    const visibility = showHeatmap ? 'visible' : 'none';
    if (map.current.getLayer('air-quality-heat')) {
      map.current.setLayoutProperty('air-quality-heat', 'visibility', visibility);
    }
    if (map.current.getLayer('intensity-background')) {
      map.current.setLayoutProperty('intensity-background', 'visibility', visibility);
    }
  }, [showHeatmap]);

  // Gerçek verileri sanal verilerle genişlet - daha yoğun bir ısı haritası için
  function expandAirQualityData(data) {
    const expandedData = [...data];
    
    // Her gerçek veri noktası için etrafına birkaç sanal veri noktası ekle
    data.forEach(location => {
      const [lng, lat] = decodeGeohash(location.geohash);
      
      // Mevcut veri noktasının etrafına rastgele noktalar ekle
      for (let i = 0; i < 5; i++) {
        const offsetLng = lng + (Math.random() - 0.5) * 5;
        const offsetLat = lat + (Math.random() - 0.5) * 5;
        
        // Metrik değerlerini biraz rastgele değiştir
        const variance = 0.8 + Math.random() * 0.4; // 0.8 ile 1.2 arası
        
        expandedData.push({
          ...location,
          geohash: `virtual_${i}_${location.geohash}`,
          // Sanal nokta için koordinatları doğrudan sakla
          virtual_lng: offsetLng,
          virtual_lat: offsetLat,
          pm25_avg: location.pm25_avg * variance,
          pm10_avg: location.pm10_avg * variance,
          no2_avg: location.no2_avg * variance,
          so2_avg: location.so2_avg * variance,
          o3_avg: location.o3_avg * variance
        });
      }
    });
    
    return expandedData;
  }

  // Geohash decoder veya sanal koordinat alıcı
function decodeGeohash(geohash) {
    try {
      const { lat, lon } = Geohash.decode(geohash);
      return [lon, lat]; // Mapbox [longitude, latitude] formatını bekler
    } catch (error) {
      console.error('Geohash çözme hatası:', error);
      return [0, 0]; // Hata durumunda varsayılan konum
    }
  }

  return (
    <div className="app">
      <div className="header">
        <div className="logo-title">
        <img src={`${process.env.PUBLIC_URL}/logo512.png`} alt="Hava Kalitesi Logosu" className="app-logo" />
          <h1>Air Quality Monitoring</h1>
        </div>
        <div className="controls">
          <div className="metric-selector">
            <label htmlFor="metric-select">Gösterilen Metrik:</label>
            <select 
              id="metric-select"
              value={selectedMetric} 
              onChange={(e) => setSelectedMetric(e.target.value)}
            >
              <option value="pm25_avg">PM2.5</option>
              <option value="pm10_avg">PM10</option>
              <option value="no2_avg">NO₂</option>
              <option value="so2_avg">SO₂</option>
              <option value="o3_avg">O₃</option>
            </select>
          </div>
          
          <div className="heatmap-toggle">
            <label htmlFor="heatmap-toggle">Isı Haritası:</label>
            <div className="toggle-switch">
              <input
                id="heatmap-toggle"
                type="checkbox"
                checked={showHeatmap}
                onChange={() => setShowHeatmap(!showHeatmap)}
              />
              <span className="toggle-slider"></span>
            </div>
          </div>

          {showHeatmap && (
            <div className="intensity-control">
              <label htmlFor="intensity-slider">Yoğunluk:</label>
              <input
                id="intensity-slider"
                type="range"
                min="0.5"
                max="3"
                step="0.1"
                value={heatmapIntensity}
                onChange={(e) => setHeatmapIntensity(parseFloat(e.target.value))}
              />
            </div>
          )}
          
          {/* Veri güncelleme kontrolleri */}
          <div className="update-section">
            <div className="auto-update-toggle">
              <label htmlFor="auto-update-toggle">Otomatik Güncelleme:</label>
              <div className="toggle-switch">
                <input
                  id="auto-update-toggle"
                  type="checkbox"
                  checked={autoUpdate}
                  onChange={() => setAutoUpdate(!autoUpdate)}
                />
                <span className="toggle-slider"></span>
              </div>
            </div>
            
            <div className="interval-select">
              <label htmlFor="update-interval">Güncelleme Sıklığı:</label>
              <select
                id="update-interval"
                value={updateInterval}
                onChange={(e) => setUpdateInterval(parseInt(e.target.value))}
                disabled={!autoUpdate}
              >
                <option value="10">10 saniye</option>
                <option value="30">30 saniye</option>
                <option value="60">1 dakika</option>
                <option value="300">5 dakika</option>
                <option value="600">10 dakika</option>
                <option value="1800">30 dakika</option>
              </select>
            </div>
            
            <button 
              className="manual-update-btn"
              onClick={fetchData}
              disabled={loading}
            >
              {loading ? 'Güncelleniyor...' : 'Şimdi Güncelle'}
            </button>
          </div>
        </div>
      </div>
      
      <div className="map-container" ref={mapContainer}>
        {/* Son güncelleme bilgisi */}
        {lastUpdated && (
          <div className="last-updated-info">
            <span className="update-indicator"></span>
            Son güncelleme: {lastUpdated.toLocaleTimeString()}
            {autoUpdate && (
              <span className="next-update-info">
                Sonraki güncelleme: {new Date(lastUpdated.getTime() + updateInterval * 1000).toLocaleTimeString()}
              </span>
            )}
          </div>
        )}
      </div>
      
      {/* Diğer bileşenler (legend, loading, error) */}
      
      {loading && <div className="loading">Veriler yükleniyor...</div>}
      {error && <div className="error">{error}</div>}
    </div>
  );
}

export default App;
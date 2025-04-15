// App.js
import Geohash from 'latlon-geohash'; // Geohash kütüphanesi
import React, { useState, useEffect, useRef } from 'react';
import mapboxgl from 'mapbox-gl';
import axios from 'axios';
import 'mapbox-gl/dist/mapbox-gl.css';
import './App.css';
// Mapbox'ın ısı haritası için gerekli eklenti
// Mapbox'ın kendi ısı haritası kullanıldığından bu import gerekli değil
// import { HeatmapLayer } from 'react-map-gl';

// Mapbox token - Gerçek uygulamada .env dosyasında saklanmalıdır
mapboxgl.accessToken = 'pk.eyJ1IjoibXVrZWxlcyIsImEiOiJjbTlpOGRiazcwMDF3MmtzZDUzc3VvZ3k1In0.yKyyKGnzLeQU9kreNrw8MA';

function App() {
  const mapContainer = useRef(null);
  const map = useRef(null);
  const [airQualityData, setAirQualityData] = useState([]);
  const [selectedMetric, setSelectedMetric] = useState('pm25_avg');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showHeatmap, setShowHeatmap] = useState(false); // Isı haritasını gösterme durumu

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

  // API'den verileri çekme
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await axios.get('http://localhost:5000/regional-averages');
        setAirQualityData(response.data);
        setLoading(false);
      } catch (err) {
        setError('Veriler yüklenirken hata oluştu.');
        setLoading(false);
        console.error('Veri çekme hatası:', err);
      }
    };

    fetchData();
  }, []);

  // Haritayı başlatma
  useEffect(() => {
    if (map.current) return; // Haritayı tekrar başlatmayı önle
    
    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/standard-satellite',
      center: [30.7133, 40.7667], // Kocaeli'nin yakınına başlangıç
      zoom: 3 // Global görünüm için zoom seviyesi
    });

    map.current.addControl(new mapboxgl.NavigationControl(), 'top-right');
    
    // Harita yüklendiğinde
    map.current.on('load', () => {
      // Ana veri kaynağı - nokta gösterimi için
      map.current.addSource('air-quality-source', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: []
        }
      });

      // Isı haritası için ikinci veri kaynağı
      map.current.addSource('air-quality-heatmap-source', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: []
        }
      });

      // Isı haritası katmanı ekleme
      map.current.addLayer({
        id: 'air-quality-heatmap',
        type: 'heatmap',
        source: 'air-quality-heatmap-source',
        layout: {
          visibility: 'none' // Başlangıçta gizli
        },
        paint: {
          // Isı yoğunluğu - metriğe göre değişecek
          'heatmap-weight': [
            'interpolate', ['linear'], ['get', 'weight'],
            0, 0,
            100, 1
          ],
          // Zoom seviyesine göre ısı yoğunluğu
          'heatmap-intensity': [
            'interpolate', ['linear'], ['zoom'],
            0, 1,
            9, 3
          ],
          // Renk dağılımı
          'heatmap-color': [
            'interpolate', ['linear'], ['heatmap-density'],
            0, 'rgba(0, 228, 0, 0)',
            0.2, 'rgba(0, 228, 0, 0.7)',
            0.4, 'rgba(255, 255, 0, 0.7)',
            0.6, 'rgba(255, 126, 0, 0.7)',
            0.8, 'rgba(255, 0, 0, 0.7)',
            1, 'rgba(153, 0, 76, 0.7)'
          ],
          // Radius zoom ile değişiyor
          'heatmap-radius': [
            'interpolate', ['linear'], ['zoom'],
            0, 15,
            9, 40
          ],
          // Yakınlaştıkça şeffaflaşma
          'heatmap-opacity': [
            'interpolate', ['linear'], ['zoom'],
            7, 1,
            9, 0.7
          ]
        }
      });

      // Nokta katmanı
      map.current.addLayer({
        id: 'air-quality-circles',
        type: 'circle',
        source: 'air-quality-source',
        paint: {
          'circle-radius': [
            'interpolate', ['linear'], ['zoom'],
            3, 8,  // Zoom seviyesi 3'te boyut 8px
            8, 16  // Zoom seviyesi 8'de boyut 16px
          ],
          'circle-color': ['get', 'color'],
          'circle-opacity': 0.8,
          'circle-stroke-width': 1,
          'circle-stroke-color': '#fff'
        }
      });

      // Tooltip için popup
      const popup = new mapboxgl.Popup({
        closeButton: false,
        closeOnClick: false
      });

      map.current.on('mouseenter', 'air-quality-circles', (e) => {
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
          <p><strong>Zaman aralığı:</strong> ${new Date(properties.start_time).toLocaleDateString()} - ${new Date(properties.end_time).toLocaleDateString()}</p>
        `;

        popup.setLngLat(coordinates)
          .setHTML(html)
          .addTo(map.current);
      });

      map.current.on('mouseleave', 'air-quality-circles', () => {
        map.current.getCanvas().style.cursor = '';
        popup.remove();
      });
    });
  }, []);

  // Harita verilerini güncelleme
  useEffect(() => {
    if (!map.current || !map.current.isStyleLoaded() || airQualityData.length === 0) return;

    // GeoJSON veri yapısı oluşturma
    const geojsonData = {
      type: 'FeatureCollection',
      features: airQualityData.map(location => {
        // Geohash'ten enlem ve boylam koordinatları çıkarma (basitleştirilmiş)
        // Gerçek uygulamada daha doğru bir geohash çözücü kullanılmalıdır
        const [lng, lat] = decodeGeohash(location.geohash);
        
        return {
          type: 'Feature',
          geometry: {
            type: 'Point',
            coordinates: [lng, lat]
          },
          properties: {
            ...location,
            color: getColor(location[selectedMetric], selectedMetric)
          }
        };
      })
    };

    // Isı haritası için GeoJSON veri yapısı
    const heatmapData = {
      type: 'FeatureCollection',
      features: airQualityData.map(location => {
        const [lng, lat] = decodeGeohash(location.geohash);
        
        // Metrik değerini 0-100 arasında normalize et
        // Bu değerler metriğe göre ayarlanabilir
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
            weight: normalizedValue
          }
        };
      })
    };

    // Nokta veri kaynağını güncelle
    if (map.current.getSource('air-quality-source')) {
      map.current.getSource('air-quality-source').setData(geojsonData);
    }
    
    // Isı haritası veri kaynağını güncelle
    if (map.current.getSource('air-quality-heatmap-source')) {
      map.current.getSource('air-quality-heatmap-source').setData(heatmapData);
    }
  }, [airQualityData, selectedMetric]);
  
  // Isı haritasını göster/gizle
  useEffect(() => {
    if (!map.current || !map.current.isStyleLoaded()) return;
    
    const visibility = showHeatmap ? 'visible' : 'none';
    if (map.current.getLayer('air-quality-heatmap')) {
      map.current.setLayoutProperty('air-quality-heatmap', 'visibility', visibility);
    }
  }, [showHeatmap]);

  // Basit bir geohash dekoder (gerçek uygulamada daha doğru bir kütüphane kullanın)
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
          <h1>Air Quality Monitor</h1>
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
        </div>
      </div>
      
      <div className="map-container" ref={mapContainer} />
      
      <div className="legend">
        <h3>Hava Kalitesi Göstergesi</h3>
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: '#00e400' }}></span>
          <span>İyi</span>
        </div>
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: '#ffff00' }}></span>
          <span>Orta</span>
        </div>
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: '#ff7e00' }}></span>
          <span>Hassas Gruplar İçin Sağlıksız</span>
        </div>
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: '#ff0000' }}></span>
          <span>Sağlıksız</span>
        </div>
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: '#99004c' }}></span>
          <span>Çok Sağlıksız</span>
        </div>
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: '#7e0023' }}></span>
          <span>Tehlikeli</span>
        </div>
      </div>
      
      {loading && <div className="loading">Veriler yükleniyor...</div>}
      {error && <div className="error">{error}</div>}
    </div>
  );
}


export default App;
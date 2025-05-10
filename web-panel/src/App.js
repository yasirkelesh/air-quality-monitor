// App.js
import React, { useState, useEffect, useRef, useCallback } from 'react';
import mapboxgl from 'mapbox-gl';
import axios from 'axios';
import Geohash from 'latlon-geohash';
import * as turf from '@turf/turf';
import 'mapbox-gl/dist/mapbox-gl.css';
import LocationAnalysisPanel from './LocationAnalysisPanel'; 
import './App.css';

// Mapbox token
mapboxgl.accessToken = 'pk.eyJ1IjoibXVrZWxlcyIsImEiOiJjbTlpOGRiazcwMDF3MmtzZDUzc3VvZ3k1In0.yKyyKGnzLeQU9kreNrw8MA';
const SSE_URL = '/sse-events';
const ANOMALIES_URL = '/anomalies'; // Anomali servisi URL'si
const REGIONAL_AVERAGES_URL = '/regional-averages'; // Bölgesel ortalama servisi URL'si
const NOTIFICATION_URL = '/notification'; // Bildirim servisi URL'si

function App() {
  const mapContainer = useRef(null);
  const map = useRef(null);
  const [airQualityData, setAirQualityData] = useState([]);
  const [selectedMetric, setSelectedMetric] = useState('pm25_avg');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showHeatmap, setShowHeatmap] = useState(true);
  const [heatmapIntensity, setHeatmapIntensity] = useState(1.5);
  const [autoUpdate, setAutoUpdate] = useState(false);
  const [updateInterval, setUpdateInterval] = useState(60);
  const [lastUpdated, setLastUpdated] = useState(null);
  const intervalRef = useRef(null);
  const [showAnomalySummary, setShowAnomalySummary] = useState(false);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [selectedLocation, setSelectedLocation] = useState(null);
  
  // Anomali takibi için yeni state
  const [anomalies, setAnomalies] = useState([]);
  const [sseStatus, setSseStatus] = useState('connecting'); // 'connecting', 'connected', 'error', 'closed'
  const anomalyTimeouts = useRef({});
  const anomalyEventSource = useRef(null);
   // Anomali servisi URL'si
  // SSE bağlantısı için useEffect

  const [notificationForm, setNotificationForm] = useState({
    email: '',
    city: '',
    geohash: ''
  });
  const [notificationStatus, setNotificationStatus] = useState(null);
  const [isNotificationRegistered, setIsNotificationRegistered] = useState(() => {
    // Sayfa yüklendiğinde localStorage'dan bildirim durumunu kontrol et
    const savedStatus = localStorage.getItem('notificationRegistered');
    return savedStatus === 'true';
  });

  useEffect(() => {
    console.log('SSE bağlantısı başlatılıyor...');
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;

    const connectToAnomalyStream = () => {
      // Önceki bağlantıyı kapat
      if (anomalyEventSource.current) {
        anomalyEventSource.current.close();
      }

      // SSE bağlantısı kur
      try {
        console.log('SSE bağlantısı kuruluyor...');
        setSseStatus('connecting');
        
        anomalyEventSource.current = new EventSource(SSE_URL);
        
        anomalyEventSource.current.onopen = () => {
          console.log('SSE bağlantısı açıldı');
          setSseStatus('connected');
          reconnectAttempts = 0;
        };
        
        anomalyEventSource.current.onmessage = (event) => {
          try {
            const anomalyData = JSON.parse(event.data);
            console.log('Yeni anomali alındı:', anomalyData);
            
            // Anomaliyi listeye ekle
            setAnomalies(prevAnomalies => [...prevAnomalies, anomalyData]);
            
            // 1 saat sonra anomaliyi kaldır
            anomalyTimeouts.current[anomalyData.id] = setTimeout(() => {
              setAnomalies(prevAnomalies => 
                prevAnomalies.filter(anomaly => anomaly.id !== anomalyData.id)
              );
              delete anomalyTimeouts.current[anomalyData.id];
            }, 3600000); // 1 saat = 3600000 ms
          } catch (parseError) {
            console.error('Anomali verisi parse edilemedi:', parseError);
          }
        };

        anomalyEventSource.current.onerror = (error) => {
          console.error('SSE Bağlantı hatası:', error);
          setSseStatus('error');
          anomalyEventSource.current.close();
          
          // Yeniden bağlanma mantığı
          if (reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++;
            const delay = Math.min(5000 * reconnectAttempts, 30000); // Max 30 saniye
            console.log(`Yeniden bağlanma denemesi ${reconnectAttempts} / ${maxReconnectAttempts}. ${delay/1000} saniye sonra...`);
            setTimeout(connectToAnomalyStream, delay);
          } else {
            console.error('SSE bağlantısı kurulamadı. Maksimum deneme aşıldı.');
            setSseStatus('closed');
          }
        };
        
      } catch (error) {
        console.error('SSE bağlantısı başlatılamadı:', error);
        setSseStatus('error');
      }
    };

    connectToAnomalyStream();

    // Cleanup
    return () => {
      if (anomalyEventSource.current) {
        anomalyEventSource.current.close();
      }
      // Tüm timeout'ları temizle
      const timeouts = anomalyTimeouts.current;
      Object.values(timeouts).forEach(clearTimeout);
    };
  }, []);

  // Anomali çemberlerini haritaya eklemek için useEffect
// Anomali çemberlerini haritaya eklemek için useEffect
useEffect(() => {
  if (!map.current) return; 
  
  // Harita hazır değilse bekle
  if (!map.current.isStyleLoaded()) {
    console.log('Harita henüz yüklenmedi, çember çizimi erteleniyor...');
    
    // Harita hazır olduğunda yeniden dene
    const checkMapAndDrawCircles = () => {
      if (map.current && map.current.isStyleLoaded()) {
        console.log('Harita hazır, çemberler çiziliyor...');
        drawAnomalyCircles();
      } else {
        console.log('Harita hala hazır değil, tekrar deneniyor...');
        setTimeout(checkMapAndDrawCircles, 500);
      }
    };
    
    checkMapAndDrawCircles();
    return;
  }
  
  // Çember çizme fonksiyonu
  function drawAnomalyCircles() {
    console.log(`Çemberler çiziliyor... (${anomalies.length} anomali)`);
    
    // Anomali verisini güncelle - her anomali için 25km yarıçaplı çember oluştur
    const anomalyFeatures = anomalies.flatMap(anomaly => {
      // Merkez noktası (anomali pozisyonu)
      const centerPoint = {
        type: 'Feature',
        geometry: {
          type: 'Point',
          coordinates: [anomaly.longitude, anomaly.latitude]
        },
        properties: {
          id: anomaly.id,
          description: anomaly.description,
          pollutant: anomaly.pollutant,
          current_value: anomaly.current_value,
          average_value: anomaly.average_value,
          increase_ratio: anomaly.increase_ratio,
          timestamp: anomaly.timestamp,
          type: 'center'
        }
      };

      // 25km yarıçaplı çember (polygon olarak)
      const circle = turf.circle(
        [anomaly.longitude, anomaly.latitude], 
        25, // 25 km yarıçap
        {
          steps: 64, // Daha yumuşak bir çember için
          units: 'kilometers'
        }
      );

      circle.properties = {
        id: anomaly.id,
        type: 'area'
      };

      return [centerPoint, circle];
    });

    // Anomali kaynağını güncelle veya oluştur
    if (map.current.getSource('anomaly-data')) {
      console.log('Mevcut anomali kaynağı güncelleniyor...');
      map.current.getSource('anomaly-data').setData({
        type: 'FeatureCollection',
        features: anomalyFeatures
      });
    } else {
      console.log('Yeni anomali kaynağı oluşturuluyor...');
      
      // İlk kez oluşturuluyorsa kaynak ve katmanları ekle
      try {
        map.current.addSource('anomaly-data', {
          type: 'geojson',
          data: {
            type: 'FeatureCollection',
            features: anomalyFeatures
          }
        });

        // 25km'lik çemberler için layer ekle
        map.current.addLayer({
          id: 'anomaly-circles',
          type: 'fill',
          source: 'anomaly-data',
          filter: ['==', ['get', 'type'], 'area'],
          paint: {
            'fill-color': 'rgba(255, 0, 0, 0.2)',
            'fill-opacity': 0.5
          }
        });

        // Çember sınırları için kontur
        map.current.addLayer({
          id: 'anomaly-circles-outline',
          type: 'line',
          source: 'anomaly-data',
          filter: ['==', ['get', 'type'], 'area'],
          paint: {
            'line-color': 'rgba(255, 0, 0, 0.8)',
            'line-width': 2
          }
        });

        // Anomali merkezleri için nokta işaretçileri
        map.current.addLayer({
          id: 'anomaly-centers',
          type: 'circle',
          source: 'anomaly-data',
          filter: ['==', ['get', 'type'], 'center'],
          paint: {
            'circle-radius': 5,
            'circle-color': '#ff0000',
            'circle-stroke-width': 2,
            'circle-stroke-color': '#ffffff'
          }
        });

        // Anomali popup
        const anomalyPopup = new mapboxgl.Popup({
          closeButton: false,
          closeOnClick: false
        });

        map.current.on('mouseenter', 'anomaly-centers', (e) => {
          map.current.getCanvas().style.cursor = 'pointer';
          
          const coordinates = e.features[0].geometry.coordinates.slice();
          const properties = e.features[0].properties;
          
          const time = new Date(properties.timestamp).toLocaleString();
          const html = `
            <div class="anomaly-popup">
              <h3><strong>⚠️ Anomali Tespit Edildi</strong></h3>
              <p><strong>Kirletici:</strong> ${properties.pollutant.toUpperCase()}</p>
              <p><strong>Değer:</strong> ${properties.current_value.toFixed(2)} µg/m³</p>
              <p><strong>Normal Ortalama:</strong> ${properties.average_value} µg/m³</p>
              <p><strong>Artış Oranı:</strong> %${(properties.increase_ratio * 100).toFixed(1)}</p>
              <p><strong>Zaman:</strong> ${time}</p>
              <p><strong>Etki Alanı:</strong> 25km yarıçaplı</p>
              <p class="anomaly-description">${properties.description}</p>
            </div>
          `;

          anomalyPopup.setLngLat(coordinates)
            .setHTML(html)
            .addTo(map.current);
        });

        map.current.on('mouseleave', 'anomaly-centers', () => {
          map.current.getCanvas().style.cursor = '';
          anomalyPopup.remove();
        });
        
        console.log('Anomali katmanları başarıyla eklendi');
      } catch (error) {
        console.error('Anomali çemberlerini oluştururken hata:', error);
      }
    }
  }
  
  // Çember çizme fonksiyonunu çağır
  drawAnomalyCircles();
  
}, [anomalies]);
  // Veri çekme fonksiyonu - useCallback ile memoize ediliyor
  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const response = await axios.get(REGIONAL_AVERAGES_URL);
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


  const fetchExistingAnomalies = useCallback(async () => {
    try {
      console.log('Mevcut anomaliler yükleniyor...');
      const response = await axios.get(ANOMALIES_URL);
      
      if (response.data && response.data.anomalies && Array.isArray(response.data.anomalies)) {
        console.log(`${response.data.count} adet mevcut anomali bulundu`);
        
        // Verileri SSE formatına dönüştür ve state'e ekle
        const formattedAnomalies = response.data.anomalies.map(anomaly => ({
          id: anomaly._id,
          source: anomaly.source,
          timestamp: anomaly.timestamp,
          anomaly_type: anomaly.anomaly_type,
          pollutant: anomaly.pollutant,
          current_value: anomaly.current_value,
          average_value: anomaly.average_value,
          increase_ratio: anomaly.increase_ratio,
          geohash: anomaly.geohash,
          geohash_prefix: anomaly.geohash_prefix,
          // Koordinatları doğrudan alınabilir formata çevir
          longitude: anomaly.location.coordinates[0],
          latitude: anomaly.location.coordinates[1],
          country: anomaly.country,
          city: anomaly.city,
          district: anomaly.district,
          description: anomaly.description,
          detected_at: anomaly.detected_at,
          expiry_time: anomaly.expiry_time
        }));
        
        setAnomalies(formattedAnomalies);
        
        // Her anomali için sona erme zamanına göre otomatik kaldırma 
        formattedAnomalies.forEach(anomaly => {
          const expiryTime = new Date(anomaly.expiry_time);
          const now = new Date();
          
          // Sona erme süresi gelecekte mi kontrol et
          if (expiryTime > now) {
            const timeoutDuration = expiryTime.getTime() - now.getTime();
            console.log(`Anomali ${anomaly.id} için ${timeoutDuration}ms sonra kaldırılacak`);
            
            anomalyTimeouts.current[anomaly.id] = setTimeout(() => {
              setAnomalies(prevAnomalies => 
                prevAnomalies.filter(a => a.id !== anomaly.id)
              );
              delete anomalyTimeouts.current[anomaly.id];
            }, timeoutDuration);
          }
        });
      }
    } catch (err) {
      console.error('Mevcut anomalileri yükleme hatası:', err);
    }
  }, []);

  // İlk veri yüklemesi
  useEffect(() => {
    fetchData();
    
    // Harita yüklendikten sonra anomalileri çekmeyi garantiye almak için
    // Hem ilk yüklemede hem de harita hazır olduğunda anomalileri çek
    fetchExistingAnomalies();
    
    // Harita hazır olduğunda ayrıca kontrol et (geç yüklenen haritalar için)
    if (map.current) {
      map.current.on('load', () => {
        console.log('Harita yüklendi, anomaliler yeniden çekiliyor...');
        fetchExistingAnomalies();
      });
    }
  }, [fetchData, fetchExistingAnomalies]);

  // Otomatik güncelleme zamanlayıcısı
  useEffect(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    
    if (autoUpdate) {
      intervalRef.current = setInterval(() => {
        fetchData();
      }, updateInterval * 1000);
      
      console.log(`Otomatik güncelleme aktif: ${updateInterval} saniye aralıkla`);
    }
    
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [autoUpdate, updateInterval, fetchData]);

  // ... (Mevcut kodun geri kalanı aynı kalacak)
  
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

      map.current.on('click', (e) => {
        // Tıklanan noktanın koordinatlarını al
        const coordinates = [e.lngLat.lng, e.lngLat.lat];
        
        // Seçilen konumu güncelle
        setSelectedLocation(coordinates);
        
        // Paneli aç
        setIsPanelOpen(true);
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
  const expandAirQualityData = useCallback((data) => {
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
  }, []);

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

  // Şehir bilgisinden geohash oluşturan fonksiyon
  const getGeohashForCity = (city) => {
    // Türkiye'deki bazı şehirlerin koordinatları
    const cityCoordinates = {
      'İstanbul': { lat: 41.0082, lon: 28.9784 },
      'Ankara': { lat: 39.9334, lon: 32.8597 },
      'İzmir': { lat: 38.4237, lon: 27.1428 },
      'Bursa': { lat: 40.1885, lon: 29.0610 },
      'Antalya': { lat: 36.8841, lon: 30.7056 },
      'Adana': { lat: 37.0000, lon: 35.3213 },
      'Konya': { lat: 37.8667, lon: 32.4833 },
      'Gaziantep': { lat: 37.0662, lon: 37.3833 },
      'Şanlıurfa': { lat: 37.1591, lon: 38.7969 },
      'Kocaeli': { lat: 40.8533, lon: 29.8815 },
      'Mersin': { lat: 36.8000, lon: 34.6333 },
      'Diyarbakır': { lat: 37.9144, lon: 40.2306 },
      'Kayseri': { lat: 38.7312, lon: 35.4787 },
      'Eskişehir': { lat: 39.7767, lon: 30.5206 },
      'Samsun': { lat: 41.2867, lon: 36.3300 },
      'Denizli': { lat: 37.7765, lon: 29.0864 },
      'Malatya': { lat: 38.3552, lon: 38.3095 },
      'Sivas': { lat: 39.7477, lon: 37.0179 },
      'Erzurum': { lat: 39.9000, lon: 41.2700 },
      'Van': { lat: 38.4891, lon: 43.4089 }
    };

    // Şehir adını normalize et (Türkçe karakterleri düzelt)
    const normalizedCity = city
      .toLowerCase()
      .replace(/ı/g, 'i')
      .replace(/ğ/g, 'g')
      .replace(/ü/g, 'u')
      .replace(/ş/g, 's')
      .replace(/ö/g, 'o')
      .replace(/ç/g, 'c');

    // Şehir koordinatlarını bul
    const cityKey = Object.keys(cityCoordinates).find(key => 
      key.toLowerCase()
        .replace(/ı/g, 'i')
        .replace(/ğ/g, 'g')
        .replace(/ü/g, 'u')
        .replace(/ş/g, 's')
        .replace(/ö/g, 'o')
        .replace(/ç/g, 'c') === normalizedCity
    );

    if (cityKey) {
      const coords = cityCoordinates[cityKey];
      return Geohash.encode(coords.lat, coords.lon, 3); // 3 karakterlik geohash
    }

    // Eğer şehir bulunamazsa varsayılan olarak İstanbul'un geohash'ini döndür
    return 'sz0';
  };

  // Bildirim kayıt fonksiyonu
  const handleNotificationSubmit = async (e) => {
    e.preventDefault();
    try {
      const geohash = getGeohashForCity(notificationForm.city);
      const formData = {
        email: notificationForm.email,
        city: notificationForm.city,
        geohash: geohash
      };

      console.log('Gönderilen veri:', formData);
      const response = await axios.post(NOTIFICATION_URL, formData);
      setNotificationStatus('success');
      setNotificationForm({ email: '', city: '', geohash: '' });
      setIsNotificationRegistered(true);
      
      // Başarılı kayıt durumunu localStorage'a kaydet
      localStorage.setItem('notificationRegistered', 'true');
      // E-posta adresini de saklayalım (opsiyonel)
      localStorage.setItem('registeredEmail', formData.email);
    } catch (error) {
      console.error('Bildirim kayıt hatası:', error);
      setNotificationStatus('error');
    }
  };

  // Bildirim durumunu sıfırlama fonksiyonu (opsiyonel)
  const resetNotificationStatus = () => {
    setIsNotificationRegistered(false);
    localStorage.removeItem('notificationRegistered');
    localStorage.removeItem('registeredEmail');
  };

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
      
      <div className="map-container" ref={mapContainer} >
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

        {/* Anomali bildirim bölümü */}
        {anomalies.length > 0 && (
          <div className="anomaly-notification">
            <div className="anomaly-header">
              <span className="anomaly-icon">⚠️</span>
              <span>Aktif Anomaliler: {anomalies.length}</span>
            </div>
          </div>
        )}

        {/* Anomali tablosu toggle butonu */}
        {anomalies.length > 0 && !showAnomalySummary && (
          <div 
            className="anomaly-summary-toggle"
            onClick={() => setShowAnomalySummary(true)}
          >
            <span className="anomaly-icon">⚠️</span>
            <span>Anomali Listesini Göster ({anomalies.length})</span>
          </div>
        )}

        {/* Anomali özet tablosu */}
        {showAnomalySummary && (
          <div className="anomaly-summary">
            <h3>
              <span className="anomaly-icon">⚠️</span>
              <span>Aktif Anomaliler ({anomalies.length})</span>
              <button 
                style={{ marginLeft: 'auto', background: 'none', border: 'none', color: 'white', cursor: 'pointer' }}
                onClick={() => setShowAnomalySummary(false)}
              >
                ✕
              </button>
            </h3>
            
            {anomalies.length > 0 ? (
              <table className="anomaly-summary-table">
                <thead>
                  <tr>
                    <th>Kirletici</th>
                    <th>Değer</th>
                    <th>Artış</th>
                    <th>Konum</th>
                  </tr>
                </thead>
                <tbody>
                  {anomalies.map(anomaly => (
                    <tr key={anomaly.id}>
                      <td>
                        <span className={`pollutant-badge pollutant-${anomaly.pollutant}`}>
                          {anomaly.pollutant.toUpperCase()}
                        </span>
                      </td>
                      <td className={anomaly.increase_ratio > 3 ? 'critical' : ''}>
                        {anomaly.current_value.toFixed(1)} µg/m³
                      </td>
                      <td>
                        {(anomaly.increase_ratio * 100).toFixed(0)}%
                      </td>
                      <td>
                        {anomaly.district || anomaly.city || anomaly.country || 'Bilinmiyor'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p>Aktif anomali bulunmuyor.</p>
            )}
          </div>
        )}
        {/* Konum analiz paneli */}
        <LocationAnalysisPanel 
          isOpen={isPanelOpen}
          onClose={() => setIsPanelOpen(false)}
          selectedLocation={selectedLocation}
          airQualityData={airQualityData}
          selectedMetric={selectedMetric}
        />

        {/* SSE bağlantı durumu göstergesi */}
        <div className={`sse-status ${sseStatus}`}>
          <span className="status-indicator"></span>
          <span className="status-text">
            {sseStatus === 'connecting' && 'Anomali servisi bağlanıyor...'}
            {sseStatus === 'connected' && 'Anomali servisi aktif'}
            {sseStatus === 'error' && 'Bağlantı hatası'}
            {sseStatus === 'closed' && 'Bağlantı kapalı'}
          </span>
        </div>

        {/* Anomali bildirim bölümü */}
        {anomalies.length > 0 && (
          <div className="anomaly-notification">
            <div className="anomaly-header">
              <span className="anomaly-icon">⚠️</span>
              <span>Aktif Anomaliler: {anomalies.length}</span>
            </div>
          </div>
        )}

        {/* Bildirim kayıt formu */}
        {!isNotificationRegistered ? (
          <div className="notification-form">
            <h3>Hava Kalitesi Bildirimleri</h3>
            <form onSubmit={handleNotificationSubmit}>
              <div className="form-group">
                <input
                  type="email"
                  placeholder="E-posta adresiniz"
                  value={notificationForm.email}
                  onChange={(e) => setNotificationForm({...notificationForm, email: e.target.value})}
                  required
                />
              </div>
              <div className="form-group">
                <input
                  type="text"
                  placeholder="Şehir"
                  value={notificationForm.city}
                  onChange={(e) => setNotificationForm({...notificationForm, city: e.target.value})}
                  required
                />
              </div>
              <button type="submit">Bildirimleri Aktifleştir</button>
            </form>
            {notificationStatus === 'success' && (
              <div className="success-message">
                Bildirimler başarıyla aktifleştirildi!
              </div>
            )}
            {notificationStatus === 'error' && (
              <div className="error-message">
                Bir hata oluştu. Lütfen tekrar deneyin.
              </div>
            )}
          </div>
        ) : (
          <div className="notification-success">
            <div className="success-icon">✓</div>
            <p>Bildirimler aktif!</p>
            <p className="notification-info">
              {localStorage.getItem('registeredEmail')} adresine hava kalitesi anomali durumlarında bilgilendirme yapılacaktır.
            </p>
            {/* Opsiyonel: Bildirimleri iptal etme butonu */}
            <button 
              className="cancel-notification-btn"
              onClick={resetNotificationStatus}
            >
              Bildirimleri İptal Et
            </button>
          </div>
        )}
      </div>
      
      {/* Diğer bileşenler (legend, loading, error) */}
      
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

      {loading && <div className="loading">Veriler yükleniyor...</div> }
      {error && <div className="error">{error}</div>}
    </div>
  );
}

export default App;
#!/bin/bash

# auto-test.sh - API veri toplayıcı test betiği
# Kullanım: ./auto-test.sh [options]
# Opsiyonlar:
#   --duration=<seconds>: Script'in çalışma süresi
#   --rate=<requests_per_second>: Saniyede kaç istek atılacağı  
#   --anomaly-chance=<percentage>: Anomali oluşturma olasılığı (0-100)

# Varsayılan değerler
DURATION=60  # saniye cinsinden
RATE=1       # saniyede istek sayısı
ANOMALY_CHANCE=10  # yüzde (0-100)
API_URL="http://localhost:8000/api/data-collector/api/v1/pollution"

# Parametreleri işle
for i in "$@"; do
  case $i in
    --duration=*)
      DURATION="${i#*=}"
      ;;
    --rate=*)
      RATE="${i#*=}"
      ;;
    --anomaly-chance=*)
      ANOMALY_CHANCE="${i#*=}"
      ;;
    *)
      echo "Bilinmeyen parametre: $i"
      exit 1
      ;;
  esac
done

# İstanbul için yaklaşık koordinat sınırları
LAT_MIN=40.8
LAT_MAX=41.2
LON_MIN=28.5
LON_MAX=29.5

# Normal değer aralıkları
PM25_MIN=5
PM25_MAX=25
PM10_MIN=10
PM10_MAX=50
NO2_MIN=10
NO2_MAX=40
SO2_MIN=5
SO2_MAX=20
O3_MIN=20
O3_MAX=60

# Anomali değer aralıkları
PM25_ANOMALY_MIN=50
PM25_ANOMALY_MAX=200
PM10_ANOMALY_MIN=100
PM10_ANOMALY_MAX=300
NO2_ANOMALY_MIN=80
NO2_ANOMALY_MAX=150
SO2_ANOMALY_MIN=40
SO2_ANOMALY_MAX=100
O3_ANOMALY_MIN=100
O3_ANOMALY_MAX=180

# Rasgele sayı üretme fonksiyonu (min ve max arasında)
random_float() {
  min=$1
  max=$2
  precision=$3
  echo "scale=$precision; $min + ($max - $min) * $RANDOM / 32767" | bc
}

# Rasgele konum üretme fonksiyonu
generate_random_location() {
  lat=$(random_float $LAT_MIN $LAT_MAX 4)
  lon=$(random_float $LON_MIN $LON_MAX 4)
  echo "$lat $lon"
}

# Rasgele kirlilik değeri üretme fonksiyonu
generate_pollution_data() {
  is_anomaly=$1
  
  if [ "$is_anomaly" = true ]; then
    pm25=$(random_float $PM25_ANOMALY_MIN $PM25_ANOMALY_MAX 1)
    pm10=$(random_float $PM10_ANOMALY_MIN $PM10_ANOMALY_MAX 1)
    no2=$(random_float $NO2_ANOMALY_MIN $NO2_ANOMALY_MAX 1)
    so2=$(random_float $SO2_ANOMALY_MIN $SO2_ANOMALY_MAX 1)
    o3=$(random_float $O3_ANOMALY_MIN $O3_ANOMALY_MAX 1)
    echo "ANOMALİ veri gönderiliyor: PM2.5=$pm25, PM10=$pm10"
  else
    pm25=$(random_float $PM25_MIN $PM25_MAX 1)
    pm10=$(random_float $PM10_MIN $PM10_MAX 1)
    no2=$(random_float $NO2_MIN $NO2_MAX 1)
    so2=$(random_float $SO2_MIN $SO2_MAX 1)
    o3=$(random_float $O3_MIN $O3_MAX 1)
  fi
  
  echo "$pm25 $pm10 $no2 $so2 $o3"
}

# Anomali durumu kontrolü
is_anomaly() {
  random=$((RANDOM % 100 + 1))
  [ $random -le $ANOMALY_CHANCE ]
}

# API isteği gönderme fonksiyonu
send_request() {
  location=$(generate_random_location)
  lat=$(echo $location | cut -d' ' -f1)
  lon=$(echo $location | cut -d' ' -f2)
  
  anomaly=false
  if is_anomaly; then
    anomaly=true
  fi
  
  pollution=$(generate_pollution_data $anomaly)
  pm25=$(echo $pollution | cut -d' ' -f1)
  pm10=$(echo $pollution | cut -d' ' -f2)
  no2=$(echo $pollution | cut -d' ' -f3)
  so2=$(echo $pollution | cut -d' ' -f4)
  o3=$(echo $pollution | cut -d' ' -f5)
  
  json_data="{\"latitude\": $lat, \"longitude\": $lon, \"pm25\": $pm25, \"pm10\": $pm10, \"no2\": $no2, \"so2\": $so2, \"o3\": $o3}"
  
  curl -s -X POST $API_URL \
    -H "Content-Type: application/json" \
    -d "$json_data" > /dev/null
    
  echo "[$(date +"%H:%M:%S")] Veri gönderildi: lat=$lat, lon=$lon, pm25=$pm25, pm10=$pm10, no2=$no2, so2=$so2, o3=$o3"
}

# Ana döngü
echo "Test başlıyor: Süre=$DURATION saniye, Hız=$RATE istek/saniye, Anomali Şansı=$ANOMALY_CHANCE%"
echo "Hedef API: $API_URL"
echo "------------------------------------------------------"

start_time=$(date +%s)
end_time=$((start_time + DURATION))
request_count=0

while [ $(date +%s) -lt $end_time ]; do
  for ((i=1; i<=$RATE; i++)); do
    send_request
    request_count=$((request_count + 1))
  done
  
  # Rate limiti için bekleme
  if [ $RATE -lt 10 ]; then
    sleep 1
  else
    # Yüksek hızlarda çok kısa bekleme
    sleep 0.1
  fi
done

echo "------------------------------------------------------"
echo "Test tamamlandı: $request_count istek gönderildi"
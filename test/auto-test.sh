#!/bin/bash

# Otomatik test script'i
# Kullanım: ./auto-test.sh [options]

# Renk tanımlamaları
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Varsayılan değerler
DURATION=300  # 5 dakika
RATE=1        # saniyede 1 istek
ANOMALY_CHANCE=10  # %10 anomali olasılığı

# Türkiye'deki bazı şehirlerin koordinatları
declare -A CITIES=(
    ["istanbul"]="41.0082,28.9784"
    ["ankara"]="39.9208,32.8541"
    ["izmir"]="38.4192,27.1287"
    ["antalya"]="36.8969,30.7133"
    ["bursa"]="40.1828,29.0663"
)

# Parametreleri işle
for arg in "$@"; do
    case $arg in
        --duration=*)
            DURATION="${arg#*=}"
            ;;
        --rate=*)
            RATE="${arg#*=}"
            ;;
        --anomaly-chance=*)
            ANOMALY_CHANCE="${arg#*=}"
            ;;
        *)
            echo -e "${RED}Bilinmeyen parametre: $arg${NC}"
            exit 1
            ;;
    esac
done

# Rastgele değer üret
random_value() {
    local min=$1
    local max=$2
    echo "$(awk -v min=$min -v max=$max 'BEGIN{srand(); print min+rand()*(max-min)}')"
}

# Rastgele şehir seç
random_city() {
    local cities=("${!CITIES[@]}")
    local random_index=$((RANDOM % ${#cities[@]}))
    echo "${cities[$random_index]}"
}

# Anomali değeri üret
generate_anomaly() {
    local base_value=$1
    local multiplier=$(random_value 2 5)
    echo "$(awk -v base=$base_value -v mult=$multiplier 'BEGIN{print base*mult}')"
}

# Veri gönder
send_data() {
    local city=$1
    local coords=(${CITIES[$city]//,/ })
    local lat=${coords[0]}
    local lon=${coords[1]}
    
    # Anomali kontrolü
    local is_anomaly=$((RANDOM % 100 < ANOMALY_CHANCE))
    
    # Değerleri oluştur
    local pm25=$(random_value 10 30)
    local pm10=$(random_value 15 35)
    local no2=$(random_value 20 40)
    local so2=$(random_value 5 25)
    local o3=$(random_value 30 50)
    
    # Anomali varsa değerleri yükselt
    if [ $is_anomaly -eq 1 ]; then
        pm25=$(generate_anomaly $pm25)
        pm10=$(generate_anomaly $pm10)
    fi
    
    # JSON verisi oluştur
    local json_data=$(cat <<EOF
{
    "sensor_id": "auto_test_$(date +%s)",
    "location": {
        "latitude": $lat,
        "longitude": $lon
    },
    "parameters": {
        "pm25": $pm25,
        "pm10": $pm10,
        "no2": $no2,
        "so2": $so2,
        "o3": $o3
    }
}
EOF
)
    
    # API'ye veri gönder
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$json_data" \
        http://localhost:8000/api/data-collector/api/v1/pollution > /dev/null
    
    if [ $? -eq 0 ]; then
        if [ $is_anomaly -eq 1 ]; then
            echo -e "${YELLOW}Anomali verisi gönderildi: $city${NC}"
        else
            echo -e "${GREEN}Veri gönderildi: $city${NC}"
        fi
    else
        echo -e "${RED}Veri gönderilirken hata oluştu${NC}"
    fi
}

# Ana döngü
echo -e "${GREEN}Test başlatılıyor...${NC}"
echo "Süre: $DURATION saniye"
echo "Hız: $RATE istek/saniye"
echo "Anomali olasılığı: %$ANOMALY_CHANCE"

start_time=$(date +%s)
end_time=$((start_time + DURATION))

while [ $(date +%s) -lt $end_time ]; do
    for ((i=0; i<RATE; i++)); do
        city=$(random_city)
        send_data "$city"
    done
    sleep 1
done

echo -e "${GREEN}Test tamamlandı${NC}"
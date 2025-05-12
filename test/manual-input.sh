#!/bin/bash

# Manuel veri girişi script'i
# Kullanım: ./manual-input.sh <latitude> <longitude> <parameter> <value>

# Renk tanımlamaları
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Parametre kontrolü
if [ "$#" -ne 4 ]; then
    echo -e "${RED}Hata: Yanlış parametre sayısı${NC}"
    echo "Kullanım: ./manual-input.sh <latitude> <longitude> <parameter> <value>"
    echo "Örnek: ./manual-input.sh 41.0082 28.9784 pm25 35.5"
    exit 1
fi

# Parametreleri al
LATITUDE=$1
LONGITUDE=$2
PARAMETER=$3
VALUE=$4

# Parametre doğrulama
if ! [[ "$LATITUDE" =~ ^-?[0-9]+\.?[0-9]*$ ]] || ! [[ "$LONGITUDE" =~ ^-?[0-9]+\.?[0-9]*$ ]]; then
    echo -e "${RED}Hata: Geçersiz koordinat değerleri${NC}"
    exit 1
fi

if ! [[ "$VALUE" =~ ^[0-9]+\.?[0-9]*$ ]]; then
    echo -e "${RED}Hata: Geçersiz değer${NC}"
    exit 1
fi

# Parametre adını kontrol et
case $PARAMETER in
    "pm25"|"pm10"|"no2"|"so2"|"o3")
        ;;
    *)
        echo -e "${RED}Hata: Geçersiz parametre adı${NC}"
        echo "Geçerli parametreler: pm25, pm10, no2, so2, o3"
        exit 1
        ;;
esac

# JSON verisi oluştur
JSON_DATA=$(cat <<EOF
{
    "sensor_id": "manual_input_$(date +%s)",
    "location": {
        "latitude": $LATITUDE,
        "longitude": $LONGITUDE
    },
    "parameters": {
        "$PARAMETER": $VALUE
    }
}
EOF
)

# API'ye veri gönder
echo -e "${GREEN}Veri gönderiliyor...${NC}"
echo "$JSON_DATA"

RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$JSON_DATA" \
    http://localhost:8000/api/data-collector/api/v1/pollution)

# Yanıtı kontrol et
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Veri başarıyla gönderildi${NC}"
    echo "Yanıt: $RESPONSE"
else
    echo -e "${RED}Veri gönderilirken hata oluştu${NC}"
    exit 1
fi
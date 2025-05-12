
#!/bin/bash

# manual-input.sh - Kirlilik parametrelerini manuel olarak girmek için betik
# Kullanım: ./manual-input.sh <latitude> <longitude> <parameter> <value>
# Örnek: ./manual-input.sh 41.0082 28.9784 pm25 35.7

API_URL="http://localhost:8000/api/data-collector/api/v1/pollution"

# Parametreleri kontrol et
if [ $# -ne 4 ]; then
    echo "Hata: Eksik parametre!"
    echo "Kullanım: ./manual-input.sh <latitude> <longitude> <parameter> <value>"
    echo "Örnek: ./manual-input.sh 41.0082 28.9784 pm25 35.7"
    echo ""
    echo "Geçerli kirlilik parametreleri: pm25, pm10, no2, so2, o3"
    exit 1
fi

LATITUDE=$1
LONGITUDE=$2
PARAMETER=$(echo "$3" | tr '[:upper:]' '[:lower:]')  # Parametreyi küçük harfe çevir
VALUE=$4

# Latitude ve longitude değerlerinin sayı olup olmadığını kontrol et
if ! [[ $LATITUDE =~ ^[+-]?[0-9]*\.?[0-9]+$ ]]; then
    echo "Hata: Latitude bir sayı olmalıdır."
    exit 1
fi

if ! [[ $LONGITUDE =~ ^[+-]?[0-9]*\.?[0-9]+$ ]]; then
    echo "Hata: Longitude bir sayı olmalıdır."
    exit 1
fi

# Value değerinin sayı olup olmadığını kontrol et
if ! [[ $VALUE =~ ^[+-]?[0-9]*\.?[0-9]+$ ]]; then
    echo "Hata: Değer bir sayı olmalıdır."
    exit 1
fi

# Geçerli bir kirlilik parametresi olup olmadığını kontrol et
case $PARAMETER in
    pm25|pm10|no2|so2|o3)
        # Geçerli parametre
        ;;
    *)
        echo "Hata: Geçersiz kirlilik parametresi: $PARAMETER"
        echo "Geçerli kirlilik parametreleri: pm25, pm10, no2, so2, o3"
        exit 1
        ;;
esac

# Geçerli varsayılan değerler
DEFAULT_PM25=15.0
DEFAULT_PM10=30.0
DEFAULT_NO2=25.0
DEFAULT_SO2=10.0
DEFAULT_O3=40.0

# Mevcut değerleri varsayılanlarla ayarla
PM25=$DEFAULT_PM25
PM10=$DEFAULT_PM10
NO2=$DEFAULT_NO2
SO2=$DEFAULT_SO2
O3=$DEFAULT_O3

# Kullanıcı tarafından belirtilen parametreyi güncelle
case $PARAMETER in
    pm25)
        PM25=$VALUE
        ;;
    pm10)
        PM10=$VALUE
        ;;
    no2)
        NO2=$VALUE
        ;;
    so2)
        SO2=$VALUE
        ;;
    o3)
        O3=$VALUE
        ;;
esac

# JSON verisini oluştur
JSON_DATA="{\"latitude\": $LATITUDE, \"longitude\": $LONGITUDE, \"pm25\": $PM25, \"pm10\": $PM10, \"no2\": $NO2, \"so2\": $SO2, \"o3\": $O3}"

# API isteği gönder
echo "Gönderilen veri:"
echo "$JSON_DATA"
echo ""

RESPONSE=$(curl -s -X POST $API_URL \
           -H "Content-Type: application/json" \
           -d "$JSON_DATA")

# Cevabı kontrol et
if [ $? -eq 0 ]; then
    echo "Veri başarıyla gönderildi!"
    echo "API yanıtı:"
    echo "$RESPONSE"
else
    echo "Hata: Veri gönderilemedi!"
    echo "API yanıtı:"
    echo "$RESPONSE"
fi
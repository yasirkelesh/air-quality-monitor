#!/bin/bash
# rabbitmq-setup-alt.sh
# RabbitMQ Exchange, Queue ve Binding oluşturma script'i (rabbitmqadmin kullanarak)

# Yapılandırma bilgileri
RABBITMQ_HOST=${RABBITMQ_HOST:-"localhost"}
RABBITMQ_PORT=${RABBITMQ_PORT:-"15672"}
RABBITMQ_USER=${RABBITMQ_USER:-"admin"}
RABBITMQ_PASS=${RABBITMQ_PASS:-"password123"}
RABBITMQ_VHOST=${RABBITMQ_VHOST:-"/"}

EXCHANGE_NAME="pollution.data"
EXCHANGE_TYPE="topic"
QUEUE_NAME="raw-data"
ROUTING_KEY="raw.data"

# Renk tanımlamaları
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}RabbitMQ yapılandırması başlatılıyor...${NC}"

# rabbitmqadmin aracını indir
if ! command -v rabbitmqadmin &> /dev/null; then
    echo "rabbitmqadmin indiriliyor..."
    curl -s -o rabbitmqadmin http://$RABBITMQ_HOST:$RABBITMQ_PORT/cli/rabbitmqadmin
    chmod +x rabbitmqadmin
    RABBITMQADMIN="./rabbitmqadmin"
else
    RABBITMQADMIN="rabbitmqadmin"
fi

# Exchange oluştur
echo "Exchange oluşturuluyor: $EXCHANGE_NAME"
$RABBITMQADMIN -H $RABBITMQ_HOST -P $RABBITMQ_PORT -u $RABBITMQ_USER -p $RABBITMQ_PASS -V $RABBITMQ_VHOST declare exchange \
    name=$EXCHANGE_NAME type=$EXCHANGE_TYPE durable=true

# Kuyruk oluştur
echo "Kuyruk oluşturuluyor: $QUEUE_NAME"
$RABBITMQADMIN -H $RABBITMQ_HOST -P $RABBITMQ_PORT -u $RABBITMQ_USER -p $RABBITMQ_PASS -V $RABBITMQ_VHOST declare queue \
    name=$QUEUE_NAME durable=true

# Binding oluştur
echo "Binding oluşturuluyor: $EXCHANGE_NAME -> $ROUTING_KEY -> $QUEUE_NAME"
$RABBITMQADMIN -H $RABBITMQ_HOST -P $RABBITMQ_PORT -u $RABBITMQ_USER -p $RABBITMQ_PASS -V $RABBITMQ_VHOST declare binding \
    source=$EXCHANGE_NAME destination=$QUEUE_NAME routing_key=$ROUTING_KEY

echo -e "${GREEN}RabbitMQ yapılandırması tamamlandı!${NC}"
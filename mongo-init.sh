#!/bin/bash
set -e

# Mongo komutlarını oluştur
mongosh <<EOF
use admin
db.auth('$MONGO_INITDB_ROOT_USERNAME', '$MONGO_INITDB_ROOT_PASSWORD')

use $MONGO_INITDB_DATABASE

// Koleksiyonları oluştur
db.createCollection('raw_data')

// Gerekli indeksleri oluştur
db.raw_data.createIndex({ "timestamp": 1 })
db.raw_data.createIndex({ "latitude": 1, "longitude": 1 })
db.raw_data.createIndex({ "source": 1 })

// Uygulama kullanıcısını oluştur
db.createUser({
  user: '$MONGO_APP_USERNAME',
  pwd: '$MONGO_APP_PASSWORD',
  roles: [
    {
      role: 'readWrite',
      db: '$MONGO_INITDB_DATABASE'
    }
  ]
})

print('MongoDB yapılandırması tamamlandı')
EOF
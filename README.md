<div align="center">
  
![Logo](./assets/images/logo1.png)
  
</div>

## Proje Hakkında
Hava Kirliliği İzleme ve Analiz Sistemi, dünya genelindeki gerçek zamanlı hava kirliliği verilerini toplayan, işleyen, analiz eden, görselleştiren ve kullanıcılara bildirim gönderen web tabanlı bir platformdur. Sistem, sensörlerden veya kullanıcılar tarafından iletilen verileri toplayarak, belirlenen eşik değerlerini aşan kirlilik seviyelerini tespit eder ve kullanıcılara anlık uyarılar gönderir. Ayrıca, anomali tespiti yaparak olağandışı durumları saptar ve kullanıcıların verilere web arayüzü üzerinden kolayca erişmesini sağlar.

### Proje Amacı ve Kapsamı

- Sensörlerden ve Kullanıcılardan gelen verilerin toplanması 
- Veri akışının yönetilmesi
- Anomali tespiti yapılması
- API üzerinden veri erişimi
- Web tabanlı dashboard ile verilerin görselleştirilmesi

## Web Panel

![web-panel](./assets/images/web-panel.png)

### Gösterge Paneli
- **Gösterilen Metrik:** metrik seçimi yapılabilir
- **Isı Haritası:** Açma/kapama ve yoğunluk ayarı
- **Otomatik Güncelleme:** Açma/kapama ve sıklık ayarı 

### Anomali Takibi
- **Aktif Anomaliler:** Mevcut anomali sayısı görüntüleme
- **Anomali Servisi:** Etkinleştirme/devre dışı bırakma seçeneği
- **Detaylar:** Kirletici türü, konsantrasyon ve eşik değerleri

### Harita Özellikleri
- **Konum İşaretleri:** İzleme istasyonları sarı noktalarla gösterilir
- **Seçili Konum:** Detaylı ölçüm verileri (PM2.5, PM10, NO2, SO2, O3)
- **Zaman Aralığı:** Ölçüm tarih aralığı gösterimi

### Hava Kalitesi Göstergesi
- İyi (Yeşil), Orta (Sarı), Hassas Gruplar İçin Sağlıksız (Turuncu)
- Sağlıksız (Kırmızı), Çok Sağlıksız (Mor), Tehlikeli (Koyu Mor)

### E-posta Bildirim Sistemi

Kullanıcılar web paneli üzerindeki formu doldurarak hava kalitesi bildirimlerine abone olabilirler. Sistem şu anda sadece Türkiye'deki şehirler için şehir ölçeğinde çalışmaktadır.

**Form Özellikleri:**
- Kullanıcı e-posta adresi girişi
- Şehir seçimi yapabilme

**Bildirim Mekanizması:**
- Kullanıcının seçtiği şehirde bir anomali tespit edildiğinde otomatik e-posta gönderimi yapılır
- E-posta içeriğinde anomali tipi, şiddeti, ölçüm değerleri ve zaman bilgisi yer alır
- Kullanıcıya sağlık önerileri de e-posta içinde sunulur

![Email Bildirim Örneği](./assets/images/mail.png)

## Sistem Mimarisi

![Sistem Mimarisi](./assets/images/architecture.png)

Platform, mikroservis mimarisi kullanılarak tasarlanmıştır ve toplam dört bağımsız servis içermektedir:

- **Data-Collector**: Kullanıcı ve sensör verilerini toplayarak veri tabanına kaydeder.
- **Data-Processing**: Toplanan verileri işleyerek analiz edilmek üzere hazırlar.
- **Anomaly-Detection**: Veriler üzerinde anomali tespiti yaparak olağandışı durumları belirler.
- **Notification (SMTP)**: Anomali veya eşik aşımı durumlarında kullanıcılara e-posta yoluyla bildirim gönderir.

Tüm sistem, Docker kullanarak konteyner tabanlı bir mimari ile yapılandırılmıştır. İletişim altyapısında RabbitMQ mesaj kuyruğu sistemi, sensör verilerinin alınmasında ise Mosquitto MQTT broker kullanılmıştır. Verilerin saklanması için ise zaman serisi verileri için InfluxDB, genel veri depolama ve kullanıcı yönetimi için MongoDB tercih edilmiştir.

Bu yapı sayesinde sistem, ölçeklenebilir, yönetilebilir ve farklı kullanım senaryolarına kolayca adapte olabilecek bir yapıya sahiptir.

## Data Collector Servisi
![Veri Toplama Katmanı](./assets/images/data-collector.png)

#### 1. Veri Kaynakları 
* **REST API İstekleri**: Manuel veri girişi için HTTP endpointleri
* **MQTT Mesajları**: Sensör verilerini almak için MQTT abonelikleri

#### 2. Sunum Katmanı (Presentation Layer)
* **HTTP Handlers**: REST API isteklerini karşılar ve işler
* **MQTT Handler**: MQTT mesajlarını dinler ve işler
* **RabbitMQ Publisher**: Kontrol edilen verinin broker'a akrarılması

#### 3. Servis Katmanı (Service Layer)
* **Data Verification**: Toplanan verilerin doğrulama işlemi yapılır

#### 4. Veri Erişim Katmanı (Data Access Layer)
* **MongoDB Repostroy**: Ham verilerin veritabanına aktarılır

## Data Processing Servisi
![Veri İşleme Katmanı](./assets/images/data-processing.png)

#### 1. Sunum Katmanı (Presentation Layer)
* **HTTP Handlers**: Bölgesel ortalama değerlerini sorgulamak için API sunar
* **RabbitMQ Consumer**: raw-data kuyruğundan ham verileri alır
* **RabbitMQ Publisher**: İşlenmiş verileri processed-data kuyruğuna gönderir

#### 2. Servis Katmanı (Service Layer)
* **Veri İşleme Servisi**: Ham verileri işler ve koordinatlardan geohash oluşturur
* **Geocoding Servisi**: Koordinatları adres bilgilerine dönüştürür
* **Bölgesel Ortalama Servisi**: Geohash bölgeleri için ortalama değerler hesaplar

#### 3. Veri Erişim Katmanı (Data Access Layer)
* **InfluxDB Repository**: Zaman serisi veritabanına erişim sağlar
* **Cache Repository**: Sık kullanılan verileri önbellekler (opsiyonel)

#### Geohash Hassasiyeti
* **5 Karakterli Geohash**: Yaklaşık 4.9 km²'lik bir alanı temsil eder
* **3 Karakterli Geohash**: Yaklaşık 156 km²'lik bir alanı temsil eder (25-30 km yarıçap)
* **Anomali Tespiti**: 25 km yarıçap içindeki farklılıkları kontrol etmek için 3 karakterli geohash kullanılır


## Anomaly Detection Servisi

![Anomali Tespit Katmanı](./assets/images/anomaly_detection.png)

#### 1. Sunum Katmanı (Presentation Layer)

* **HTTP Handlers**: Anomali sorguları için API sunar
* **RabbitMQ Consumer**: processed-data kuyruğundan veri alır
* **RabbitMQ Publisher**: İşlenmiş verileri anomaly-data kuyruğuna gönderir
* **Server-Sent Events**: Gerçek zamanlı anomali bildirimlerini yönetir


#### 2. Servis Katmanı (Service Layer)

* **Anomali Tespiti Servisi**: Ana orkestratör, tüm tespit tiplerini koordine eder
* **Zaman Serisi Anomali Dedektörü**: %50 artış kontrolünü yapar
* **Eşik Anomali Dedektörü**: WHO standartlarına göre kontrol yapar
* **Mekansal Anomali Dedektörü**: 25km yarıçap içindeki farklılıkları kontrol eder

#### 4. Veri Erişim Katmanı (Data Access Layer)

* **MongoDB Repository**: Mekansal Veriler ve anomali kayıtlarını yönetir

#### Anomali Tespit Kriterleri
* **Zaman Serisi Anomalisi**: Bölge ortalamasına göre %50 veya daha fazla artış
* **Mekansal Anomali**: Bölge ortalamasına göre %100 veya daha fazla artış

#### Veri Saklama ve İndeksleme
* **MongoDB TTL İndeksi**: Anomaliler 1 saat sonra otomatik olarak silinir
* **Geospatial İndeks**: Harita bazlı sorgular için konum indeksi
* **Benzersiz İndeks**: Aynı anomali tekrar kaydedilmez

## Notification Servisi

![Notification Servisi](./assets/images/notification-service.png)

#### 1. Sunum Katmanı (Presentation Layer)

* **RabbitMQ Consumer**: anomaly-data kuyruğundan veri alır
* **HTTP Handlers**: Bildirim Kayıtları için API'ler sağlar

#### 2. Servis Katmanı (Service Layer)

* **Abonelik Yönetim Servisi**: Abonelik toplanması ve takibi servisi
* **Bildirim Servisi**: Bildirim olşturma durumlarını kontrol eder
* **Email Gönderme Servisi**: SMTP server kullanrak kullanıcıya Email bildirmleri gönderilir


#### 4. Veri Erişim Katmanı (Data Access Layer)

* **MongoDB Repository**: Kullanıcı Email adresieri ve gönderilen mailleri yönetir

## API Gateway

![Api Gateway](./assets/images/api-gateway.png)

Bu API Gateway mimarisi, Go diliyle geliştirilmiş olup, gelen kullanıcı isteklerini Middleware (Auth & Logger) katmanında işler ve Proxy üzerinden ilgili mikroservislere yönlendirir. Yapı, güvenlik (kimlik doğrulama), loglama ve yönlendirme işlevlerini modüler olarak ayrıştırarak, servislerin yalıtılmış ve yönetilebilir olmasını sağlar.


## Teknoloji Seçimleri ve Gerekçeleri

![Anomali Tespit Katmanı](./assets/images/tech_stack.png)


## Frontend


#### - React
**Neden seçildi?**: React, component tabanlı yapısıyla büyük ve ölçeklenebilir arayüzler geliştirmeye çok uygundur. Ayrıca geniş bir ekosisteme (kütüphaneler, araçlar) sahiptir ve performansı yüksektir. Modern, interaktif UI'lar oluşturmak için ideal.

#### - Mapbox
**Neden seçildi?**: Harita tabanlı uygulamalarda esnek özelleştirme ve yüksek performans sunduğu için tercih edilir. Google Maps'e göre daha fazla özgürlük ve kendi sunucunda barındırma imkanı verir.

## Backend

#### - Python (FastAPI)
**Neden seçildi?**: Python, güçlü veri işleme, yapay zeka ve analitik kütüphanelerine (NumPy, Pandas, Scikit-learn) sahiptir. FastAPI ise modern, hızlı (high-performance) API'ler geliştirmek için idealdir. Asenkron programlama destekler ve otomatik OpenAPI/Swagger belgeleri sunar.

#### - Golang (Gin)
**Neden seçildi?**: Go dili, yüksek performans, düşük bellek kullanımı ve eşzamanlılık (concurrency) desteği ile öne çıkar. Gin framework'ü ise Go ile hızlı, minimal ve verimli REST API'ler geliştirmek için kullanılır. Yüksek trafik altında bile düşük gecikme süresi sağlar.

## Database

#### - MongoDB
**Neden seçildi?**: NoSQL yapısıyla esnek ve hızlıdır. Yapılandırılmamış veya dinamik veriler için uygundur (örneğin: IoT verileri, kullanıcı aktiviteleri). JSON benzeri veri yapısıyla esnek veri modeli sunar.

#### - InfluxDB
**Neden seçildi?**: Zaman serisi verileri (time-series data) için optimize edilmiştir. Sensör verileri, anlık ölçümler, izleme (monitoring) gibi sürekli zaman bazlı veri akışlarının saklanması ve sorgulanması için idealdir.

## Messaging

#### - RabbitMQ
**Neden seçildi?**: Güvenilir mesaj kuyruğu (message queue) sistemi sağlar. Mikroservisler arası iletişimi güvenli ve asenkron hale getirir. Dağıtık sistemlerde veri kaybı olmadan iletişim kurmak için tercih edilir.

#### - MQTT
**Neden seçildi?**: Çok hafif bir mesajlaşma protokolüdür. Özellikle düşük bant genişliğine ve sınırlı kaynaklara sahip IoT cihazları arasında veri iletimi için idealdir. Pub/Sub (yayınla/abone ol) modelini kullanır.

## DevOps

#### - Docker
**Neden seçildi?**: Uygulamaların bağımlılıklarıyla birlikte paketlenip her ortamda çalıştırılabilmesini sağlar. Geliştirme, test ve üretim ortamlarında tutarlılık sunar.

#### - Docker Compose
**Neden seçildi?**: Birden fazla konteyneri (örneğin: backend, database, message broker) kolayca tanımlayıp birlikte çalıştırmak için kullanılır. Çok servislere sahip projelerde yönetimi kolaylaştırır.

#### - Nginx
**Neden seçildi?**: Yüksek performanslı bir ters proxy (reverse proxy) ve yük dengeleyicidir (load balancer). Aynı zamanda statik dosya sunucusu olarak da mükemmeldir. API gateway veya frontend/backend servislerini dış dünyaya açarken güvenli ve verimli bir katman sunar.




## Kurulum

### Önemli Not
Bu projede `.env` dosyaları ve konfigürasyon dosyaları repository'de bulunmaktadır. Bu, geliştirme ortamında kolaylık sağlamak amacıyla yapılmıştır. Ancak, gerçek bir üretim ortamında bu yaklaşım güvenlik riskleri oluşturabilir. Üretim ortamında:

* Hassas bilgileri (API anahtarları, veritabanı şifreleri vb.) `.env` dosyalarında saklamak yerine güvenli bir secrets management sistemi kullanılmalıdır
* Konfigürasyon dosyaları environment-specific olmalı ve sadece gerekli ortamda bulunmalıdır
* Tüm hassas bilgiler şifrelenmiş olarak saklanmalıdır

### Ön Koşullar

- Docker ve Docker Compose
- Git

### Adım Adım Kurulum

#### 1. Projeyi Klonlama

```bash
git clone git@github.com:yasirkelesh/global-hava-kalitesi.git
cd global-hava-kalitesi
```

#### 2. Ortam Değişkenlerini Ayarlama

Her servis için gerekli `.env` dosyalarını oluşturun:
*Not: Şu anda tek bir `.env` dosyası bulunmaktadır. Servisler dağıtıldığında ayrı `.env` dosyaları oluşturulabilir.*

**Data Collector için:**
```bash
# MongoDB
MONGODB_URI=mongodb://admin:password@mongodb:27017/
MONGODB_DATABASE=pollution_db
MONGODB_COLLECTION=raw_data

# MQTT
MQTT_BROKER_URL=mqtt://mqtt-broker:1883
MQTT_CLIENT_ID=data-collector
MQTT_TOPIC=pollution

# RabbitMQ
RABBITMQ_URI=amqp://admin:password123@rabbitmq:5672/
RABBITMQ_EXCHANGE=pollution.data
RABBITMQ_QUEUE=raw-data
RABBITMQ_ROUTING_KEY=raw.data
```

**Data Processing için:**
```bash
# RabbitMQ
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=admin
RABBITMQ_PASS=password123
RABBITMQ_EXCHANGE=pollution.data
RABBITMQ_RAW_QUEUE=raw-data
RABBITMQ_PROCESSED_QUEUE=processed-data

# InfluxDB
INFLUXDB_HOST=influxdb
INFLUXDB_PORT=8086
INFLUXDB_USERNAME=admin
INFLUXDB_PASSWORD=0YEQiQN6Y7UyU2N
INFLUXDB_BUCKET=data_processing
INFLUXDB_ORG=myorg
INFLUXDB_TOKEN=mytoken

# Geocoding
GEOHASH_PRECISION=5
```

**Anomaly Detection için:**
```bash
# RabbitMQ
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=admin
RABBITMQ_PASS=password123
RABBITMQ_EXCHANGE=pollution.data
RABBITMQ_ANOMALY_QUEUE=anomaly-data
RABBITMQ_PROCESSED_QUEUE=processed-data

# MongoDB
MONGODB_URI=mongodb://admin:password@mongodb:27017/data-collector?authSource=admin
MONGODB_DB=anomaly_db
MONGODB_COLLECTION=anomalies

# Anomali Ayarları
ANOMALY_TTL_HOURS=1
```

#### 3. MQTT Konfigürasyonu

Mosquitto broker için konfigürasyon dosyasını oluşturun:

```bash
mkdir -p mosquitto/config
echo "listener 1883
allow_anonymous true" > mosquitto/config/mosquitto.conf
```

#### 4. Servis Konfigürasyonları

**API Gateway (api-gateway/config.yaml):**
```yaml
server:
  port: "8000"
  mode: "debug"

services:
  datacollector: "http://data-collector:8080"

auth:
  jwtsecret: "your-secret-key"
  enabled: false
```

**Data Collector (data-collector/config.yaml):**
```yaml
server:
  port: "8080"
  mode: "debug"

mongodb:
  uri: "mongodb://admin:password@mongodb:27017/"
  database: "pollution_db"
  collection: "raw_data"

mqtt:
  brokerurl: "mqtt://mqtt-broker:1883"
  clientid: "data-collector"
  topic: "pollution"

rabbitmq:
  uri: "amqp://admin:password123@rabbitmq:5672/"
  exchange: "pollution.data"
  queue: "raw-data"
  routingkey: "raw.data"
```

#### 5. Sistemi Başlatma

Docker Compose ile tüm servisleri başlatın:

```bash
docker-compose up --build -d
```

Tüm container'ların çalıştığını kontrol edin:

```bash
docker-compose ps
```

#### 6. Servis Sağlık Kontrolü

Her servisin sağlık durumunu kontrol edin:

```bash
# API Gateway
curl http://localhost:8000/health

# Data Collector
curl http://localhost:8080/health

# Data Processing
curl http://localhost:5000/health

# Anomaly Detection
curl http://localhost:6000/health
```

#### 7. İlk Kurulum Sonrası İşlemler




## API Dokümantasyonu

### Data Collector Servisi

| Endpoint | Metod | Açıklama | İstek Gövdesi | Yetkilendirme |
|----------|-------|----------|--------------|---------------|
| /api/data-collector/api/v1/pollution | GET | Hava kirliliği verilerini listeler | - | Bearer Token(jwt) |
| /api/data-collector/api/v1/pollution | POST | Yeni hava kirliliği verisi ekler | `{ "sensor_id": "string", "location": { "latitude": number, "longitude": number }, "parameters": { "pm25": number, "pm10": number, "o3": number } }` | Bearer Token(jwt) |
| /api/data-collector/api/v1/health | GET | Servis sağlık kontrolü | - | - |
| /api/data-collector/ping | GET | Basit ping kontrolü | - | - |

### Data Processing Servisi

| Endpoint | Metod | Açıklama | İstek Gövdesi | Yetkilendirme |
|----------|-------|----------|--------------|---------------|
| /api/data-processing/regional-averages | GET | Tüm bölgeler için ortalama değerleri getirir | - | Bearer Token(jwt) |
| /api/data-processing/regional-average/{geohash} | GET | Belirli bir bölge için ortalama değerleri getirir | - | Bearer Token(jwt) |

### Anomaly Detection Servisi

| Endpoint | Metod | Açıklama | İstek Gövdesi | Yetkilendirme |
|----------|-------|----------|--------------|---------------|
| /api/anomaly-detection/anomalies | GET | Tespit edilen anomalileri listeler | - | Bearer Token(jwt) |
| /api/anomaly-detection/anomalies/{id} | GET | Belirli bir anomali detayını getirir | - | Bearer Token(jwt) |
| /api/anomaly-detection/anomalies/region/{geohash} | GET | Belirli bir bölgedeki anomalileri listeler | - | Bearer Token(jwt) |

### Notification Servisi

| Endpoint | Metod | Açıklama | İstek Gövdesi | Yetkilendirme |
|----------|-------|----------|--------------|---------------|
| /api/notification/api/v1/users | POST | Yeni kullanıcı oluşturur | `{ "email": "string", "city": "string", "notification_types": ["string"] }` | Bearer Token(jwt) |
| /api/notification/api/v1/health | GET | Servis sağlık kontrolü | - | - |
| /api/notification/ping | GET | Basit ping kontrolü | - | - |

### API Gateway

| Endpoint | Metod | Açıklama | İstek Gövdesi | Yetkilendirme |
|----------|-------|----------|--------------|---------------|
| /health | GET | Gateway sağlık kontrolü | - | - |
| /auth/login | POST | Kimlik doğrulama | `{ "username": "string", "password": "string" }` | - |

### Kimlik Doğrulama

API'ye erişim için JWT tabanlı kimlik doğrulama kullanılmaktadır. Token almak için:

```http
POST /auth/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

Yanıt:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2023-12-31T23:59:59Z"
}
```

Sonraki tüm isteklerde bu token'ı Authorization header'ında kullanın:
```http
GET /api/data
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```
```

## Script'lerin Kullanımı

*[Bu bölümde sistemde bulunan script'lerin kullanımını açıklayın]*

### Veri Simülasyonu

Sensör verilerini simüle etmek için:

```bash
cd scripts
python simulate_sensors.py --count 5 --interval 10
```

Parametreler:
- `--count`: Simüle edilecek sensör sayısı
- `--interval`: Veri gönderme aralığı (saniye)

### Veritabanı Bakımı

MongoDB indekslerini optimize etmek için:

```bash
python scripts/optimize_mongodb.py
```

InfluxDB retention policy'lerini yapılandırmak için:

```bash
python scripts/configure_influxdb_retention.py --days 90
```

### Sistem Sağlık Kontrolü

Tüm servislerin durumunu kontrol etmek için:

```bash
bash scripts/health_check.sh
```

## Test Senaryoları

Sistemin farklı bileşenlerini test etmek için çeşitli test senaryoları bulunmaktadır.

### MQTT Test Aracı

MQTT test aracı, sensör verilerini simüle etmek ve sistemin farklı senaryolardaki davranışını test etmek için kullanılır.

#### Kurulum

```bash
cd test
go mod tidy
```

#### Çalıştırma

```bash
go run test.go
```

Program başlatıldığında 3 farklı test senaryosu sunulur:

1. **Normal Veri Gönderimi**
   - Tek bir sensörden normal değerlerle veri gönderir
   - Her saniye bir veri gönderilir
   - PM2.5, PM10, NO2, SO2 ve O3 değerleri normal aralıklarda

2. **Anomali Testi**
   - Belirli bir sensörden anormal değerlerle veri gönderir
   - Her 10 veride bir anomali oluşturur
   - PM2.5 ve PM10 değerlerinde ani yükselişler gözlemlenir

3. **Türkiye'deki 40 Farklı Lokasyondan Veri Gönderimi**
   - Türkiye'nin 40 farklı şehrinden veri gönderir
   - Her şehir için farklı bir sensör ID'si kullanılır
   - Gerçekçi konum verileri ile test yapılır

#### Test Verisi Formatı

```json
{
    "latitude": 40.714331,
    "longitude": 29.945292,
    "timestamp": "2024-03-14T12:00:00Z",
    "pm25": 25.0,
    "pm10": 26.0,
    "no2": 24.0,
    "so2": 25.0,
    "o3": 25.5,
    "device_id": "sensor_normal"
}
```

### Manuel ve Otomatik Test Scriptleri

Sistemin farklı senaryolarda test edilmesi için iki farklı test scripti bulunmaktadır:

#### 1. Manuel Veri Girişi Script'i (`manual-input.sh`)

Belirli bir konum için kirlilik parametrelerini manuel olarak girmek için kullanılır.

**Kullanım:**
```bash
chmod +x manual-input.sh
./manual-input.sh <latitude> <longitude> <parameter> <value>
```

**Örnek:**
```bash
# PM2.5 değerini 35.7 olarak ayarla
./manual-input.sh 41.0082 28.9784 pm25 35.7

# PM10 değerini 75.2 olarak ayarla
./manual-input.sh 41.0082 28.9784 pm10 75.2

# NO2 değerini 60.5 olarak ayarla
./manual-input.sh 41.0082 28.9784 no2 60.5

# SO2 değerini 25.8 olarak ayarla
./manual-input.sh 41.0082 28.9784 so2 25.8

# O3 değerini 90.3 olarak ayarla
./manual-input.sh 41.0082 28.9784 o3 90.3
```

**Özellikler:**
- Koordinat ve parametre doğrulaması
- JSON formatında veri oluşturma
- API'ye veri gönderme
- Hata kontrolü ve renkli çıktılar

**Desteklenen Parametreler:**
- pm25
- pm10
- no2
- so2
- o3

#### 2. Otomatik Test Script'i (`auto-test.sh`)

Sistemi otomatik olarak test etmek için rastgele veriler üreten ve gönderen bir script.

**Kullanım:**
```bash
chmod +x auto-test.sh
./auto-test.sh [options]
```

**Opsiyonlar:**
- `--duration=<seconds>`: Script'in çalışma süresi (varsayılan: 300 saniye)
- `--rate=<requests_per_second>`: Saniyede kaç istek atılacağı (varsayılan: 1)
- `--anomaly-chance=<percentage>`: Anomali oluşturma olasılığı (varsayılan: %10)

**Örnekler:**
```bash
# Varsayılan ayarlarla çalıştırma
./auto-test.sh

# Kısa ama yoğun test
./auto-test.sh --duration=30 --rate=10 --anomaly-chance=5

# Uzun süreli, yavaş ve sık anomalili test
./auto-test.sh --duration=1800 --rate=1 --anomaly-chance=30
```

**Özellikler:**
- Türkiye'deki gerçek şehir koordinatları kullanır
- Rastgele ve anomali değerleri üretir
- Özelleştirilebilir test süresi, hızı ve anomali olasılığı
- Renkli çıktılar ve ilerleme gösterimi
- Hata kontrolü ve loglama

**Test Edilen Senaryolar:**
1. Normal veri akışı
2. Anomali durumları
3. Yüksek yük testi
4. Farklı lokasyonlardan eşzamanlı veri gönderimi


#### Test Sonuçlarını İzleme

1. **Data Collector Servisi**
   - `http://localhost:8000/api/data-collector/api/v1/pollution` endpoint'inden gelen verileri kontrol edin
   - MongoDB'de kaydedilen verileri kontrol edin

2. **Data Processing Servisi**
   - `http://localhost:8000/api/data-processing/regional-averages` endpoint'inden bölgesel ortalamaları kontrol edin
   - InfluxDB'de işlenmiş verileri kontrol edin

3. **Anomaly Detection Servisi**
   - `http://localhost:8000/api/anomaly-detection/anomalies` endpoint'inden tespit edilen anomalileri kontrol edin
   - SSE bağlantısı üzerinden gerçek zamanlı anomali bildirimlerini izleyin

4. **Notification Servisi**
   - E-posta bildirimlerinin gönderildiğini kontrol edin
   - MongoDB'de kaydedilen bildirim kayıtlarını kontrol edin


## Sorun Giderme Rehberi

### Yaygın Sorunlar ve Çözümleri

#### 1. Servisler Başlatılamıyor

**Belirtiler:**
- Docker container'ları başlatılamıyor
- Servisler birbirine bağlanamıyor

**Çözümler:**
```bash
# Tüm container'ları durdur ve sil
docker-compose down

# Docker volume'larını temizle
docker-compose down -v

# Yeniden başlat
docker-compose up -d

# Log'ları kontrol et
docker-compose logs -f
```

#### 2. RabbitMQ Mesaj Kuyruğu Sorunları

**Belirtiler:**
- Mesajlar işlenmiyor
- Kuyruk tıkanıklığı

**Çözümler:**
```bash
# RabbitMQ yönetim panelini aç
http://localhost:15672
# Kullanıcı adı: admin
# Şifre: password123

# Kuyruk durumunu kontrol et
docker exec -it rabbitmq rabbitmqctl list_queues

# RabbitMQ servisini yeniden başlat
docker-compose restart rabbitmq
```

### Performans İzleme

Sistem performansını izlemek için:

```bash
# Container kaynak kullanımını izle
docker stats

# Servis log'larını gerçek zamanlı izle
docker-compose logs -f [servis-adı]

# MongoDB performans metriklerini kontrol et
docker exec -it mongodb mongosh --eval "db.serverStatus()"
```


### Sistem Güncelleme

Sistemi güncellemek için:

```bash
# En son değişiklikleri çek
git pull

# Servisleri yeniden oluştur ve başlat
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```
```
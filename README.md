<div align="center">
  
![Logo](./assets/images/logo1.png)
  
</div>

## Proje Hakkında
Hava Kalitesi İzleme ve Analiz Sistemi, dünya genelindeki gerçek zamanlı hava kirliliği verilerini toplayan, işleyen, analiz eden ve görselleştiren ve kullanıcılara bildirim gönderen web tabanlı bir platformdur. Sistem, sensörlerden veya scriptler aracılığıyla iletilen verileri toplayarak, belirlenen eşik değerlerini aşan kirlilik seviyelerini tespit eder ve kullanıcılara anlık uyarılar gönderir. Ayrıca, anomali tespiti yaparak olağandışı durumları saptar ve kullanıcıların verilere web arayüzü üzerinden kolayca erişmesini sağlar.

### Proje Amacı ve Kapsamı

- Sensörlerden gelen verilerin MQTT protokolü üzerinden toplanması
- Veri akışının yönetilmesi
- Anomali tespiti yapılması
- RESTful API üzerinden veri erişimi
- Web tabanlı dashboard ile verilerin görselleştirilmesi

## Sistem Mimarisi

![Sistem Mimarisi](./assets/images/architecture.png)

Platform, mikroservis mimarisi kullanılarak tasarlanmıştır ve toplam dört bağımsız servis içermektedir:

- **Data-Collector**: MQTT broker üzerinden gelen sensör verilerini toplayarak veri tabanına kaydeder.
- **Data-Processing**: Toplanan verileri işleyerek analiz edilmek üzere hazırlar.
- **Anomaly-Detection**: Veriler üzerinde anomali tespiti yaparak olağandışı durumları belirler.
- **Notification (SMTP)**: Anomali veya eşik aşımı durumlarında kullanıcılara e-posta yoluyla bildirim gönderir.

Tüm sistem, konteyner tabanlı bir mimari ile yapılandırılmıştır. İletişim altyapısında RabbitMQ mesaj kuyruğu sistemi, sensör verilerinin alınmasında ise Mosquitto MQTT broker kullanılmıştır. Verilerin saklanması için ise zaman serisi verileri için InfluxDB, genel veri depolama ve kullanıcı yönetimi için MongoDB tercih edilmiştir.

Bu yapı sayesinde sistem, ölçeklenebilir, yönetilebilir ve farklı kullanım senaryolarına kolayca adapte olabilecek bir yapıya sahiptir.

## Data Collector Servisi
![Veri Toplama Katmanı](./assets/images/data-collector.png)

Katmanlı veri toplama servisinin mimarisi aşağıdaki bileşenlerden oluşmaktadır:

#### 1. Veri Kaynakları
* **REST API İstekleri**: Manuel veri girişi için HTTP endpointleri
* **MQTT Mesajları**: Sensör verilerini almak için MQTT abonelikleri

#### 2. Sunum Katmanı
* **HTTP Handlers**: REST API isteklerini karşılar ve işler
* **MQTT Handler**: MQTT mesajlarını dinler ve işler

#### 3. Servis Katmanı
* **Pollution Service**: Veri işleme, doğrulama ve zenginleştirme işlerini yürütür (yazma işlemleri için)
* **Query Service**: Veri sorgulama ve filtreleme işlemlerini yönetir (okuma işlemleri için)

#### 4. Altyapı Katmanı
* **MongoDB Repository**: MongoDB ile etkileşimi sağlar (hem yazma hem okuma)
* **RabbitMQ Publisher**: RabbitMQ kuyruklarına mesaj gönderimi yönetir

#### 5. Veri Katmanı
* **MongoDB**: Ham verilerin saklandığı veritabanı
* **RabbitMQ Queue**: Servisler arası iletişim için kullanılan mesaj kuyruğu

Bu katmanlı mimari, her bileşenin net bir sorumluluğa sahip olmasını ve bağımsız olarak test edilebilmesini sağlar. Ayrıca, gRPC entegrasyonu sayesinde diğer mikroservisler ve istemciler, veri toplama servisinin topladığı ham verilere verimli bir şekilde erişebilirler.
   
## Data Processing Servisi
![Veri İşleme Katmanı](./assets/images/data-processing.png)

#### 1. Sunum Katmanı (Presentation Layer)
* **RabbitMQ Consumer**: raw-data kuyruğundan ham verileri alır
* **RabbitMQ Publisher**: İşlenmiş verileri processed-data kuyruğuna gönderir
* **FastAPI**: Bölgesel ortalama değerleri sorgulamak için API sunar

#### 2. İş Katmanı (Business Layer)
* **Veri İşleme Servisi**: Ham verileri işler ve koordinatlardan geohash oluşturur
* **Geocoding Servisi**: Koordinatları adres bilgilerine dönüştürür
* **Bölgesel Ortalama Servisi**: Geohash bölgeleri için ortalama değerler hesaplar

#### 3. Veri Erişim Katmanı (Data Access Layer)
* **InfluxDB Repository**: Zaman serisi veritabanına erişim sağlar
* **Cache Repository**: Sık kullanılan verileri önbellekler (opsiyonel)

#### 4. Altyapı Katmanı (Infrastructure Layer)
* **Konfigürasyon Yönetimi**: Çevresel değişkenler ve yapılandırma ayarları
* **Loglama**: Uygulama durumlarını ve hatalarını kaydetme
* **Hata Yönetimi**: Hataları ele alma ve raporlama


## Anomaly Detection Servisi

![Anomali Tespit Katmanı](./assets/images/anomaly_detection.png)

#### 1. Sunum Katmanı (Presentation Layer)
Bu katman, dış dünya ile etkileşimi yönetir:
* **RabbitMQ Consumer**: İşlenmiş veri kuyruğundan veri alır
* **WebSocket Controller**: Gerçek zamanlı anomali bildirimlerini yönetir
* **REST API Controller**: Anomali sorguları için HTTP endpoint'leri sağlar

#### 2. İş Katmanı (Business Layer)
Bu katman, ana iş mantığını ve algoritmalarını içerir:
* **Anomali Tespiti Servisi**: Ana orkestratör, tüm tespit tiplerini koordine eder
* **Zaman Serisi Anomali Dedektörü**: %50 artış kontrolünü yapar
* **Eşik Anomali Dedektörü**: WHO standartlarına göre kontrol yapar
* **Mekansal Anomali Dedektörü**: 25km yarıçap içindeki farklılıkları kontrol eder

#### 4. Veri Erişim Katmanı (Data Access Layer)
Veritabanı ve diğer veri kaynaklarıyla etkileşimi yönetir:
* **Anomali Repository**: MongoDB'de anomali kayıtlarını yönetir
* **Mekansal Veri Repository**: Yakın sensörleri sorgulamak için


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

## - Messaging

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

*[Bu bölümde sistemin kurulum adımlarını detaylı olarak açıklayın]*

### Ön Koşullar

- Docker ve Docker Compose
- Git
- Node.js (v14 veya üzeri) *[gerekiyorsa]*
- Python 3.8+ *[gerekiyorsa]*

### Adım Adım Kurulum

#### 1. Projeyi Klonlama

```bash
git clone git@github.com:yasirkelesh/global-hava-kalitesi.git
cd hava-kalitesi-sistemi
```

#### 2. Ortam Değişkenlerini Ayarlama

`.env.example` dosyasını `.env` olarak kopyalayın ve gerekli değişkenleri düzenleyin:

```bash
cp .env.example .env
```

Düzenlemeniz gereken temel değişkenler:

```
# MongoDB
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=password
MONGO_INITDB_DATABASE=airquality

# RabbitMQ
RABBITMQ_DEFAULT_USER=admin
RABBITMQ_DEFAULT_PASS=password

# InfluxDB
INFLUXDB_USERNAME=admin
INFLUXDB_PASSWORD=password
INFLUXDB_ORG=myorg
INFLUXDB_BUCKET=airquality
INFLUXDB_ADMIN_TOKEN=my-token
```

#### 3. MQTT Konfigürasyonu

Mosquitto broker için konfigürasyon dosyasını oluşturun:

```bash
mkdir -p mosquitto/config
echo "listener 1883
allow_anonymous true" > mosquitto/config/mosquitto.conf
```

#### 4. Sistemi Başlatma

Docker Compose ile tüm servisleri başlatın:

```bash
docker-compose up -d
```

Tüm container'ların çalıştığını kontrol edin:

```bash
docker-compose ps
```

#### 5. İlk Kurulum Sonrası İşlemler

*[Gerekli ise kurulum sonrası yapılması gereken işlemleri açıklayın]*

- MongoDB için indekslerin oluşturulması
- InfluxDB bucket yapılandırması
- İlk kullanıcı hesabının oluşturulması

## Kullanım Rehberi

*[Bu bölümde sistemin nasıl kullanılacağını açıklayın]*

### Web Arayüzü Erişimi

Web arayüzüne erişmek için tarayıcınızdan şu adrese gidin:

```
http://localhost:80
```

Varsayılan giriş bilgileri:
- Kullanıcı adı: `admin`
- Şifre: `password` *[üretim ortamında değiştirin!]*

### Dashboard Kullanımı

*[Dashboard'un nasıl kullanılacağını açıklayın]*

- Ana Gösterge Paneli
- Harita Görünümü
- Zaman Serisi Grafikleri
- Raporlama Özellikleri

### Mobil Erişim

*[Eğer varsa, mobil uygulama veya mobil web arayüzü hakkında bilgi verin]*

## API Dokümantasyonu

*[Bu bölümde API'nin nasıl kullanılacağını açıklayın]*

API Gateway'e şu adresten erişebilirsiniz:

```
http://localhost:8000
```

### Kimlik Doğrulama

API'ye erişim için JWT tabanlı kimlik doğrulama kullanılmaktadır. Token almak için:

```http
POST /api/auth/login
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

### Temel Endpoint'ler

#### Anlık Veri Alma

```http
GET /api/data/current
```

Yanıt:

```json
{
  "timestamp": "2023-05-15T14:30:00Z",
  "readings": [
    {
      "sensor_id": "sensor-001",
      "location": {
        "latitude": 40.7128,
        "longitude": -74.0060
      },
      "parameters": {
        "pm25": 15.4,
        "pm10": 25.6,
        "o3": 0.034
      }
    }
  ]
}
```

#### Tarihsel Veri Sorgulama

```http
GET /api/data/history?sensor_id=sensor-001&start=2023-05-01T00:00:00Z&end=2023-05-15T23:59:59Z
```

#### Anomali Raporları

```http
GET /api/anomalies?severity=high
```

*[Diğer endpoint'leri ve örnekleri ekleyin]*

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

## Sorun Giderme (Troubleshooting)

*[Bu bölümde yaygın sorunları ve çözümlerini açıklayın]*

### Veri Akışı Sorunları

**Belirtiler:**
- Web panelde veri görünmüyor
- MQTT broker'a veri gelmiyor

**Çözümler:**
1. MQTT broker'ın çalıştığını kontrol edin:
   ```bash
   docker logs mqtt-broker
   ```

2. Data Collector servisinin loglarını kontrol edin:
   ```bash
   docker logs data-collector
   ```

3. RabbitMQ yönetim panelini kontrol edin:
   ```
   http://localhost:15672
   ```

### Veritabanı Bağlantı Sorunları

**Belirtiler:**
- API Gateway 500 Internal Server Error dönüyor
- Servislerin loglarında veritabanı bağlantı hataları görünüyor

**Çözümler:**
1. MongoDB'nin çalıştığını kontrol edin:
   ```bash
   docker logs mongodb
   ```

2. InfluxDB'nin çalıştığını kontrol edin:
   ```bash
   docker logs influxdb
   ```

3. Ağ bağlantılarını kontrol edin:
   ```bash
   docker network inspect app-network
   ```

### Web Arayüzü Sorunları

**Belirtiler:**
- Sayfalar yüklenmiyor
- Grafiklerde veri görünmüyor

**Çözümler:**
1. Web Panel container'ının çalıştığını kontrol edin:
   ```bash
   docker logs web-panel
   ```

2. API Gateway'in erişilebilir olduğunu kontrol edin:
   ```bash
   curl http://localhost:8000/health
   ```

3. Tarayıcı konsolunu kontrol edin (F12 tuşuna basın)

## Katkıda Bulunma

*[Projeye katkıda bulunmak isteyenler için bilgiler]*

1. Projeyi fork edin
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Değişikliklerinizi commit edin (`git commit -m 'Add some amazing feature'`)
4. Branch'inizi push edin (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

## Lisans

*[Lisans bilgilerini ekleyin]*

Bu proje [LİSANS ADI] altında lisanslanmıştır - detaylar için [LICENSE](LICENSE) dosyasına bakın.

## İletişim

*[İletişim bilgilerinizi ekleyin]*

Proje Sahibi - [E-posta adresiniz](mailto:email@example.com)

Proje Linki: [https://github.com/kullaniciadi/hava-kalitesi-sistemi](https://github.com/kullaniciadi/hava-kalitesi-sistemi)

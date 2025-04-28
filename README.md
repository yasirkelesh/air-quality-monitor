# Hava Kalitesi İzleme ve Analiz Sistemi

![Web-Panel](./assets/images/web-panel.png)

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

- Data-Collector: MQTT broker üzerinden gelen sensör verilerini toplayarak veri tabanına kaydeder.
- Data-Processing: Toplanan verileri işleyerek analiz edilmek üzere hazırlar.
- Anomaly-Detection: Veriler üzerinde anomali tespiti yaparak olağandışı durumları belirler.
- Notification (SMTP): Anomali veya eşik aşımı durumlarında kullanıcılara e-posta yoluyla bildirim gönderir.

Tüm sistem, konteyner tabanlı bir mimari ile yapılandırılmıştır. İletişim altyapısında RabbitMQ mesaj kuyruğu sistemi, sensör verilerinin alınmasında ise Mosquitto MQTT broker kullanılmıştır. Verilerin saklanması için ise zaman serisi verileri için InfluxDB, genel veri depolama ve kullanıcı yönetimi için MongoDB tercih edilmiştir.

Bu yapı sayesinde sistem, ölçeklenebilir, yönetilebilir ve farklı kullanım senaryolarına kolayca adapte olabilecek bir yapıya sahiptir.

### Bileşenler

1. **Veri Toplama Katmanı**
   - **MQTT Broker (Eclipse Mosquitto)**: Sensörlerden gelen verilerin toplanması için MQTT protokolünü destekler
   - **Data Collector Servisi**: Sensörlerden gelen verileri alır ve işlenmek üzere kuyruğa gönderir
![Veri Toplama Katmani](./assets/images/data-collector.png)
2. **Mesaj Kuyruk Sistemi**
   - **RabbitMQ**: Mikro servisler arasında asenkron iletişimi sağlar

3. **Veri İşleme Katmanı**
   - **Data Processing Servisi**: Ham verileri temizler, doğrular ve analiz için hazırlar

4. **Veri Depolama Katmanı**
   - **MongoDB**: Yapılandırılmış ve yarı yapılandırılmış verilerin saklanması
   - **InfluxDB**: Zaman serisi verilerinin verimli depolanması ve sorgulanması

5. **Anomali Tespiti Katmanı**
   - **Anomaly Detection Servisi**: Makine öğrenimi algoritmaları kullanarak anormal hava kalitesi değerlerini tespit eder

6. **API Katmanı**
   - **API Gateway**: Tüm servislere tek bir noktadan erişim sağlar, kimlik doğrulama ve yetkilendirme yapar

7. **Sunum Katmanı**
   - **Web Panel**: React tabanlı kullanıcı arayüzü, gösterge panelleri ve raporlamayı sağlar

### Veri Akışı

1. Sensörler MQTT protokolü üzerinden verileri Mosquitto broker'a gönderir
2. Data Collector verileri alır ve doğrular
3. Veriler RabbitMQ kuyruğuna eklenir
4. Data Processing servisi verileri işler ve MongoDB ile InfluxDB'ye kaydeder
5. Anomaly Detection servisi düzenli olarak verileri analiz eder
6. Web Panel ve API Gateway bu verilere erişim sağlar

## Teknoloji Seçimleri ve Gerekçeleri

*[Bu bölümde kullanılan teknolojileri ve neden seçildiklerini açıklayın]*

### Veri Toplama ve İletişim
- **MQTT (Eclipse Mosquitto)**: Düşük bant genişliği gerektiren, hafif bir IoT protokolü olduğu için sensör verilerini toplamada ideal
- **RabbitMQ**: Güvenilir mesaj iletimi, yük dengeleme ve servisler arası asenkron iletişim için tercih edildi

### Veritabanları
- **MongoDB**: Esnek şema yapısı sayesinde farklı sensör verilerinin depolanmasına olanak sağlar
- **InfluxDB**: Zaman serisi verileri için optimize edilmiş, yüksek performanslı sorgulamalar sunar

### Backend Servisleri
- **Node.js/Python/Go**: *[Kullandığınız dilleri ve çerçeveleri belirtin ve nedenlerini açıklayın]*
- **Docker & Docker Compose**: Kolay dağıtım ve ölçeklendirme için konteynerleştirme teknolojisi

### Frontend
- **React**: Hızlı, komponent tabanlı UI geliştirme için modern JavaScript kütüphanesi
- **Chart.js/D3.js**: *[Kullandığınız görselleştirme kütüphanelerini belirtin]*

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

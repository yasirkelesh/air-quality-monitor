Veri Toplama Servisi Mimarisi
Yukarıdaki diyagram, katmanlı veri toplama servisinin mimarisini göstermektedir. Her katmanda şu bileşenler yer alır:

1. Veri Kaynakları
REST API İstekleri: Manuel veri girişi için HTTP endpointleri

MQTT Mesajları: Sensör verilerini almak için MQTT abonelikleri

gRPC İstekleri: Ham verileri sorgulamak için gRPC endpointleri
2. Sunum Katmanı
HTTP Handlers: REST API isteklerini karşılar ve işler

MQTT Handler: MQTT mesajlarını dinler ve işler

gRPC Handler: Ham veri sorgulama isteklerini karşılar ve yanıtlar
3. Servis Katmanı
Pollution Service: Veri işleme, doğrulama ve zenginleştirme işlerini yürütür (yazma işlemleri için)

Query Service: Veri sorgulama ve filtreleme işlemlerini yönetir (okuma işlemleri için)
4. Altyapı Katmanı
MongoDB Repository: MongoDB ile etkileşimi sağlar (hem yazma hem okuma)

RabbitMQ Publisher: RabbitMQ kuyruklarına mesaj gönderimi yönetir
5. Veri Katmanı
MongoDB: Ham verilerin saklandığı veritabanı

RabbitMQ Queue: Servisler arası iletişim için kullanılan mesaj kuyruğu
Veri Akışları:
Veri Yazma Akışı:

REST/MQTT -> İlgili Handler -> Pollution Service -> MongoDB Repository & RabbitMQ Publisher -> MongoDB & RabbitMQ Queue

Veri Okuma Akışı:

gRPC İsteği -> gRPC Handler -> Reading Data -> MongoDB Repository -> MongoDB (verileri okur) -> gRPC Yanıtı
Bu katmanlı mimari, her bileşenin net bir sorumluluğa sahip olmasını ve bağımsız olarak test edilebilmesini sağlar. Ayrıca, gRPC entegrasyonu sayesinde diğer mikroservisler ve istemciler, veri toplama servisinin topladığı ham verilere verimli bir şekilde erişebilirler.
\
7tUj2EDYOYx9K2BCuvu-MPSa5BUOHBBfIydYfoexwDyilAZzz-cMXA8uMC7Me-yTE1hjXkLsDnhKzj_y0tMDMw==
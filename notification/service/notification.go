package service

import (
	"fmt"
	"log"
	"time"

	"github.com/yasirkelesh/notification/domain"
	"github.com/yasirkelesh/notification/repository"
	"gopkg.in/gomail.v2"
)

type NotificationService struct {
	userRepo   repository.UserRepository
	mailDialer *gomail.Dialer
	mailFrom   string
}

func NewNotificationService(userRepo *repository.UserRepository, smtpHost string, smtpPort int, smtpUser, smtpPass, mailFrom string) *NotificationService {
	dialer := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	return &NotificationService{
		userRepo:   *userRepo,
		mailDialer: dialer,
		mailFrom:   mailFrom,
	}
}

func (s *NotificationService) ProcessAnomaly(anomaly *domain.Anomaly) error {
	// Geohash'in ilk 3 karakterini al (bölge için)
	regionGeohash := anomaly.Geohash[:3]
	log.Printf("Bölge: %s", regionGeohash)
	// Bu bölgedeki kullanıcıları bul
	users, err := s.userRepo.FindUsersByRegion(regionGeohash)
	if err != nil {
		return fmt.Errorf("bölge kullanıcıları bulunamadı: %v", err)
	}
	log.Printf("Kullanıcılar: %v", users)
	// Her kullanıcıya bildirim gönder
	for _, user := range users {
		if err := s.sendAnomalyNotification(user, anomaly); err != nil {
			log.Printf("Kullanıcı %s için bildirim gönderilemedi: %v", user.Email, err)
			continue
		}
	}

	return nil
}

func (s *NotificationService) sendAnomalyNotification(user *domain.User, anomaly *domain.Anomaly) error {
	// E-posta konusu oluştur
	subject := fmt.Sprintf("Hava Kalitesi Anomali Uyarısı - %s", anomaly.Pollutant)

	// HTML e-posta şablonu oluştur
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hava Kalitesi Anomali Bildirimi</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.5;
            color: #e0e0e0;
            margin: 0;
            padding: 0;
            background-color: #121212;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #1e1e1e;
            padding: 20px;
            border-radius: 4px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.5);
        }
        .header {
            text-align: center;
            padding: 10px 0;
            border-bottom: 2px solid #4caf50; /* Yeşil */
        }
        .logo {
            max-height: 80px;
        }
        .content {
            padding: 15px 0;
        }
        h2 {
            color: #4caf50; /* Yeşil */
            margin-top: 0;
            text-align: center;
        }
        .details {
            background-color: #2a2a2a;
            padding: 15px;
            border-left: 4px solid #4caf50; /* Yeşil */
            margin: 15px 0;
        }
        .details-title {
            font-weight: bold;
            margin-bottom: 10px;
            color: #4caf50; /* Yeşil */
        }
        ul {
            list-style-type: none;
            padding-left: 0;
        }
        li {
            padding: 5px 0;
            border-bottom: 1px solid #333333;
        }
        li:last-child {
            border-bottom: none;
        }
        .anomaly-value {
            font-weight: bold;
            color: #4caf50; /* Yeşil */
        }
        .footer {
            text-align: center;
            padding-top: 15px;
            border-top: 1px solid #333333;
            font-size: 12px;
            color: #a0a0a0;
        }
        .cta-button {
            display: inline-block;
            background-color: #4caf50; /* Yeşil */
            color: black;
            text-decoration: none;
            padding: 10px 20px;
            border-radius: 4px;
            font-weight: bold;
            margin: 15px 0;
        }
        strong {
            color: #ffffff;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <img src="https://havakalitesi.example.com/logo.png" alt="Hava Kalitesi İzleme Sistemi Logo" class="logo">
        </div>
        <div class="content">
            <h2>Hava Kalitesi Anomali Bildirimi</h2>
            
            <p>bölgesinizde hava kalitesi anomali tespit edildi.</p>
            
            <div class="details">
                <div class="details-title">Anomali Detayları:</div>
                <ul>
                    <li><strong>Anomali Tipi:</strong> <span class="anomaly-value">%s</span></li>
                    <li><strong>Kirletici:</strong> %s</li>
                    <li><strong>Mevcut Değer:</strong> <span class="anomaly-value">%.2f</span></li>
                    <li><strong>Bölge Ortalaması:</strong> %.2f</li>
                    <li><strong>Artış Oranı:</strong> <span class="anomaly-value">%.2f</span></li>
                    <li><strong>Tarih:</strong> %s</li>
                </ul>
            </div>
            
            <p>Detaylı bilgi için web sitemizi ziyaret edebilirsiniz.</p>
            <div style="text-align: center;">
                <a href="http://localhost" class="cta-button">Web Sitesini Ziyaret Et</a>
            </div>
        </div>
        <div class="footer">
            <p>Saygılarımızla,<br>Hava Kalitesi İzleme Sistemi</p>
            <p>Bu e-posta otomatik bilgilendirme sistemi tarafından gönderilmiştir.</p>
        </div>
    </div>
</body>
</html>`,
		anomaly.AnomalyType,
		anomaly.Pollutant,
		anomaly.CurrentValue,
		anomaly.AverageValue,
		anomaly.IncreaseRatio,
		anomaly.Timestamp.Format("02.01.2006 15:04:05"))

	// Yedek olarak düz metin e-posta içeriği oluştur
	plainBody := fmt.Sprintf(`

		
		bölgesinizde hava kalitesi anomali tespit edildi.
		
		Anomali Detayları:
		- Anomali Tipi: %s
		- Kirletici: %s
		- Mevcut Değer: %.2f
		- Bölge Ortalaması: %.2f
		- Artış Oranı: %.2f
		- Tarih: %s
		
		Detaylı bilgi için web sitemizi ziyaret edebilirsiniz: https://havakalitesi.example.com
		
		Saygılarımızla,
		Hava Kalitesi İzleme Sistemi
	`, anomaly.AnomalyType, anomaly.Pollutant,
		anomaly.CurrentValue, anomaly.AverageValue, anomaly.IncreaseRatio,
		anomaly.Timestamp.Format("02.01.2006 15:04:05"))

	// E-posta oluştur
	m := gomail.NewMessage()
	m.SetHeader("From", s.mailFrom)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", plainBody)      // Düz metin için
	m.AddAlternative("text/html", htmlBody) // HTML versiyonu ekle

	// E-postayı gönder
	if err := s.mailDialer.DialAndSend(m); err != nil {
		return fmt.Errorf("e-posta gönderilemedi: %v", err)
	}

	// Bildirim kaydını oluştur
	notification := &domain.Notification{
		UserID:    user.ID,
		AnomalyID: anomaly.ID,
		Type:      "EMAIL",
		Status:    "SENT",
		SentAt:    time.Now(),
		Content:   htmlBody, // İçerik olarak HTML versiyonu kaydediyoruz
	}

	if err := s.userRepo.SaveNotification(notification); err != nil {
		log.Printf("Bildirim kaydı oluşturulamadı: %v", err)
	}

	return nil
}

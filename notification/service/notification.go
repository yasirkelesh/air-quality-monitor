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
	// E-posta şablonu oluştur
	subject := fmt.Sprintf("Hava Kalitesi Anomali Uyarısı - %s", anomaly.Pollutant)

	body := fmt.Sprintf(`
		Sayın %s,
		
		%s bölgesinde hava kalitesi anomali tespit edildi.
		
		Anomali Detayları:
		- Anomali Tipi: %s
		- Kirletici: %s
		- Mevcut Değer: %.2f
		- Bölge Ortalaması: %.2f
		- Artış Oranı: %.2f
		- Tarih: %s
		
		Detaylı bilgi için web sitemizi ziyaret edebilirsiniz.
		
		Saygılarımızla,
		Hava Kalitesi İzleme Sistemi
	`, user.Email, anomaly.Geohash, anomaly.AnomalyType, anomaly.Pollutant,
		anomaly.CurrentValue, anomaly.AverageValue, anomaly.IncreaseRatio,
		anomaly.Timestamp.Format("02.01.2006 15:04:05"))

	// E-posta oluştur
	m := gomail.NewMessage()
	m.SetHeader("From", s.mailFrom)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

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
		Content:   body,
	}

	if err := s.userRepo.SaveNotification(notification); err != nil {
		log.Printf("Bildirim kaydı oluşturulamadı: %v", err)
	}

	return nil
}

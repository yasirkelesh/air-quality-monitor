// domain/model.go (güncelleme)
package domain

import (
	"time"
)

// PollutionData hava kirliliği verisi modeli
type PollutionData struct {
	ID        string    `json:"id,omitempty" bson:"_id,omitempty"`
	Latitude  float64   `json:"latitude" binding:"required,gte=-90,lte=90" bson:"latitude"`
	Longitude float64   `json:"longitude" binding:"required,gte=-180,lte=180" bson:"longitude"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	PM25      *float64  `json:"pm25,omitempty" bson:"pm25,omitempty"`
	PM10      *float64  `json:"pm10,omitempty" bson:"pm10,omitempty"`
	NO2       *float64  `json:"no2,omitempty" bson:"no2,omitempty"`
	SO2       *float64  `json:"so2,omitempty" bson:"so2,omitempty"`
	O3        *float64  `json:"o3,omitempty" bson:"o3,omitempty"`
	Source    string    `json:"source,omitempty" bson:"source,omitempty"`
}

// NewPollutionData yeni bir PollutionData örneği oluşturur
func NewPollutionData(lat, lon float64) *PollutionData {
	return &PollutionData{
		Latitude:  lat,
		Longitude: lon,
		Timestamp: time.Now().UTC(),
		Source:    "api",
	}
}

// HasPollutants en az bir kirlilik parametresi olduğunu kontrol eder
func (p *PollutionData) HasPollutants() bool {
	return p.PM25 != nil || p.PM10 != nil || p.NO2 != nil || p.SO2 != nil || p.O3 != nil
}
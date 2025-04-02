// api/models.go
package api

// PollutionDataRequest veri alma isteği
type PollutionDataRequest struct {
	Latitude  float64  `json:"latitude" binding:"required,gte=-90,lte=90"`
	Longitude float64  `json:"longitude" binding:"required,gte=-180,lte=180"`
	Timestamp string   `json:"timestamp,omitempty"` // ISO 8601 format
	PM25      *float64 `json:"pm25,omitempty"`
	PM10      *float64 `json:"pm10,omitempty"`
	NO2       *float64 `json:"no2,omitempty"`
	SO2       *float64 `json:"so2,omitempty"`
	O3        *float64 `json:"o3,omitempty"`
	DeviceID  string   `json:"device_id,omitempty"`
}

// PollutionDataResponse başarılı yanıt
type PollutionDataResponse struct {
	Status    string `json:"status"`
	ID        string `json:"id,omitempty"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp"`
}

// ErrorResponse hata yanıtı
type ErrorResponse struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
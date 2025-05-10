package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email" binding:"required,email"`
	City      string             `bson:"city" json:"city" binding:"required"`
	Geohash   string             `bson:"geohash" json:"geohash" binding:"required"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type AnomalyType string

const (
	SpatialAnomaly    AnomalyType = "SPATIAL"
	TimeSeriesAnomaly AnomalyType = "TIME_SERIES"
)

type Anomaly struct {
	ID            string      `bson:"_id,omitempty" json:"id"`
	AnomalyType   AnomalyType `bson:"anomaly_type" json:"anomaly_type"`
	Pollutant     string      `bson:"pollutant" json:"pollutant"`
	Description   string      `bson:"description" json:"description"`
	CurrentValue  float64     `bson:"current_value" json:"current_value"`
	AverageValue  float64     `bson:"average_value" json:"average_value"`
	IncreaseRatio float64     `bson:"increase_ratio" json:"increase_ratio"`
	Geohash       string      `bson:"geohash" json:"geohash"`
	Latitude      float64     `bson:"latitude" json:"latitude"`
	Longitude     float64     `bson:"longitude" json:"longitude"`
	Timestamp     time.Time   `bson:"timestamp" json:"timestamp"`
	Source        string      `bson:"source" json:"source"`
}

type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	AnomalyID string             `bson:"anomaly_id" json:"anomaly_id"`
	Type      string             `bson:"type" json:"type"`     // EMAIL, SMS, etc.
	Status    string             `bson:"status" json:"status"` // SENT, FAILED, etc.
	SentAt    time.Time          `bson:"sent_at" json:"sent_at"`
	Content   string             `bson:"content" json:"content"`
}

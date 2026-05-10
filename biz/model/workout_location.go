package model

import (
	"time"
)

type WorkoutLocation struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	RouteID            uint      `gorm:"index;not null" json:"route_id"`
	Latitude           float64   `gorm:"not null" json:"latitude"`
	Longitude          float64   `gorm:"not null" json:"longitude"`
	Altitude           float64   `json:"altitude"`
	Timestamp          time.Time `gorm:"index" json:"timestamp"`
	HorizontalAccuracy float64   `json:"horizontal_accuracy"`
	VerticalAccuracy   float64   `json:"vertical_accuracy"`
	Speed              float64   `json:"speed"`
	Course             float64   `json:"course"`
	CreatedAt          time.Time `json:"created_at"`
}

func (WorkoutLocation) TableName() string {
	return "workout_locations"
}

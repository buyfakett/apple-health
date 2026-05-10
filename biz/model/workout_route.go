package model

import (
	"time"
)

type WorkoutRoute struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	WorkoutID  uint      `gorm:"index" json:"workout_id"`
	SourceName string    `json:"source_name"`
	StartDate  time.Time `gorm:"index" json:"start_date"`
	EndDate    time.Time `gorm:"index" json:"end_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (WorkoutRoute) TableName() string {
	return "workout_routes"
}

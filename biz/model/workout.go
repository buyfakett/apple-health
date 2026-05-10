package model

import (
	"time"
)

type Workout struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	WorkoutActivityType   string    `gorm:"index" json:"workout_activity_type"`
	Duration              float64   `json:"duration"`
	DurationUnit          string    `json:"duration_unit"`
	TotalDistance         float64   `json:"total_distance"`
	TotalDistanceUnit     string    `json:"total_distance_unit"`
	TotalEnergyBurned     float64   `json:"total_energy_burned"`
	TotalEnergyBurnedUnit string    `json:"total_energy_burned_unit"`
	SourceName            string    `json:"source_name"`
	SourceVersion         string    `json:"source_version"`
	StartDate             time.Time `gorm:"index" json:"start_date"`
	EndDate               time.Time `gorm:"index" json:"end_date"`
	CreationDate          time.Time `json:"creation_date"`
	Device                string    `json:"device"`
	Metadata              string    `gorm:"type:text" json:"metadata"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (Workout) TableName() string {
	return "workouts"
}

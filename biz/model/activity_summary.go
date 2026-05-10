package model

import (
	"time"
)

type ActivitySummary struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	DateComponents         string    `gorm:"index;not null" json:"date_components"`
	ActiveEnergyBurned     float64   `json:"active_energy_burned"`
	ActiveEnergyBurnedGoal float64   `json:"active_energy_burned_goal"`
	ActiveEnergyBurnedUnit string    `json:"active_energy_burned_unit"`
	AppleExerciseTime      float64   `json:"apple_exercise_time"`
	AppleExerciseTimeGoal  float64   `json:"apple_exercise_time_goal"`
	AppleExerciseTimeUnit  string    `json:"apple_exercise_time_unit"`
	AppleStandTime         float64   `json:"apple_stand_time"`
	AppleStandTimeGoal     float64   `json:"apple_stand_time_goal"`
	AppleStandTimeUnit     string    `json:"apple_stand_time_unit"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (ActivitySummary) TableName() string {
	return "activity_summaries"
}

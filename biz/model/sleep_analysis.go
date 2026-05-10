package model

import (
	"time"
)

type SleepAnalysis struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SourceName    string    `json:"source_name"`
	SourceVersion string    `json:"source_version"`
	Value         string    `json:"value"`
	StartDate     time.Time `gorm:"index" json:"start_date"`
	EndDate       time.Time `gorm:"index" json:"end_date"`
	CreationDate  time.Time `json:"creation_date"`
	Device        string    `json:"device"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (SleepAnalysis) TableName() string {
	return "sleep_analyses"
}

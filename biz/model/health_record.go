package model

import (
	"time"
)

type HealthRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Type           string    `gorm:"index;not null" json:"type"`
	SourceName     string    `json:"source_name"`
	SourceVersion  string    `json:"source_version"`
	Unit           string    `json:"unit"`
	Value          float64   `json:"value"`
	StartDate      time.Time `gorm:"index" json:"start_date"`
	EndDate        time.Time `gorm:"index" json:"end_date"`
	CreationDate   time.Time `json:"creation_date"`
	Device         string    `json:"device"`
	HKQuantityType string    `json:"hk_quantity_type"`
	Metadata       string    `gorm:"type:text" json:"metadata"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (HealthRecord) TableName() string {
	return "health_records"
}

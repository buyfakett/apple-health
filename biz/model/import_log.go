package model

import (
	"time"
)

type ImportLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FileName    string    `gorm:"not null" json:"file_name"`
	FileSize    int64     `json:"file_size"`
	RecordCount int       `json:"record_count"`
	Status      string    `gorm:"not null" json:"status"`
	Error       string    `gorm:"type:text" json:"error"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ImportLog) TableName() string {
	return "import_logs"
}

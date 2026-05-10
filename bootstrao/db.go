package bootstrao

import (
	"apple_health/biz/model"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.HealthRecord{},
		&model.Workout{},
		&model.WorkoutRoute{},
		&model.WorkoutLocation{},
		&model.ActivitySummary{},
		&model.SleepAnalysis{},
		&model.ImportLog{},
	); err != nil {
		return err
	}

	return nil
}

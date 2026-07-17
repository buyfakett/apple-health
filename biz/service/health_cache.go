package service

import (
	"apple_health/biz/dal"
	"apple_health/utils/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var ErrHealthCacheDisabled = errors.New("redis未配置")

// DailyHealthCache is the raw Redis value. Keep this close to the Grafana SQL output:
// one day, one step count, and that day's workout records.
type DailyHealthCache struct {
	Date           string                    `json:"date"`
	StepCount      int64                     `json:"step_count,omitempty"`
	WorkoutRecords []DailyWorkoutCacheRecord `json:"workout_records,omitempty"`
}

func (c DailyHealthCache) IsZero() bool {
	return c.StepCount == 0 && len(c.WorkoutRecords) == 0
}

type DailyWorkoutCacheRecord struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	WorkoutType     string    `json:"workout_type"`
	WorkoutName     string    `json:"workout_name"`
	DurationMinutes float64   `json:"duration_minutes,omitempty"`
	DistanceKM      float64   `json:"distance_km,omitempty"`
	EnergyKCal      float64   `json:"energy_kcal,omitempty"`
	SourceName      string    `json:"source_name"`
}

type DailyStepCount struct {
	Date      string `json:"date"`
	StepCount int64  `json:"step_count"`
}

type DailyWorkoutRecord struct {
	Date            string    `json:"date"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	TimeRange       string    `json:"time_range"`
	WorkoutType     string    `json:"workout_type"`
	WorkoutName     string    `json:"workout_name"`
	DurationMinutes float64   `json:"duration_minutes"`
	DurationText    string    `json:"duration_text"`
	DistanceKM      float64   `json:"distance_km"`
	DistanceText    string    `json:"distance_text"`
	EnergyKCal      float64   `json:"energy_kcal"`
	EnergyText      string    `json:"energy_text"`
	SourceName      string    `json:"source_name"`
}

type HealthCacheWriteSummary struct {
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	WrittenCount int      `json:"written_count"`
	WrittenDates []string `json:"written_dates"`
	SkippedCount int      `json:"skipped_count"`
	SkippedDates []string `json:"skipped_dates"`
}

type HealthCacheRange struct {
	StartDate      string               `json:"start_date"`
	EndDate        string               `json:"end_date"`
	StepCounts     []DailyStepCount     `json:"step_counts"`
	WorkoutRecords []DailyWorkoutRecord `json:"workout_records"`
}

type workoutAggregateRow struct {
	StartDate           time.Time `gorm:"column:start_date"`
	EndDate             time.Time `gorm:"column:end_date"`
	WorkoutActivityType string    `gorm:"column:workout_activity_type"`
	DurationMinutes     float64   `gorm:"column:duration_minutes"`
	DistanceKM          float64   `gorm:"column:distance_km"`
	EnergyKCal          float64   `gorm:"column:energy_kcal"`
	SourceName          string    `gorm:"column:source_name"`
}

var workoutActivityDisplayNames = map[string]string{
	"HKWorkoutActivityTypeRunning":                       "跑步",
	"HKWorkoutActivityTypeWalking":                       "步行",
	"HKWorkoutActivityTypeCycling":                       "骑行",
	"HKWorkoutActivityTypeSwimming":                      "游泳",
	"HKWorkoutActivityTypeYoga":                          "瑜伽",
	"HKWorkoutActivityTypeTraditionalStrengthTraining":   "力量训练",
	"HKWorkoutActivityTypeFunctionalStrengthTraining":    "功能性力量训练",
	"HKWorkoutActivityTypeElliptical":                    "椭圆机",
	"HKWorkoutActivityTypeRowing":                        "划船",
	"HKWorkoutActivityTypeStairs":                        "爬楼梯",
	"HKWorkoutActivityTypeHiking":                        "徒步",
	"HKWorkoutActivityTypeDance":                         "舞蹈",
	"HKWorkoutActivityTypeJumpRope":                      "跳绳",
	"HKWorkoutActivityTypeCoreTraining":                  "核心训练",
	"HKWorkoutActivityTypeFlexibility":                   "柔韧性训练",
	"HKWorkoutActivityTypeHighIntensityIntervalTraining": "高强度间歇训练",
	"HKWorkoutActivityTypeMixedCardio":                   "混合有氧",
	"HKWorkoutActivityTypeCardioDance":                   "有氧舞蹈",
	"HKWorkoutActivityTypeStairClimbing":                 "爬楼梯",
	"HKWorkoutActivityTypeDownhillSkiing":                "高山滑雪",
	"HKWorkoutActivityTypeSnowboarding":                  "单板滑雪",
	"HKWorkoutActivityTypeCrossCountrySkiing":            "越野滑雪",
	"HKWorkoutActivityTypeSurfingSports":                 "冲浪",
	"HKWorkoutActivityTypePaddleSports":                  "划桨运动",
	"HKWorkoutActivityTypeSkatingSports":                 "滑冰",
	"HKWorkoutActivityTypeBowling":                       "保龄球",
	"HKWorkoutActivityTypeGolf":                          "高尔夫",
	"HKWorkoutActivityTypeTennis":                        "网球",
	"HKWorkoutActivityTypeBadminton":                     "羽毛球",
	"HKWorkoutActivityTypeTableTennis":                   "乒乓球",
	"HKWorkoutActivityTypeBasketball":                    "篮球",
	"HKWorkoutActivityTypeSoccer":                        "足球",
	"HKWorkoutActivityTypeVolleyball":                    "排球",
	"HKWorkoutActivityTypeBaseball":                      "棒球",
	"HKWorkoutActivityTypeSoftball":                      "垒球",
	"HKWorkoutActivityTypeHandCycling":                   "手摇车",
	"HKWorkoutActivityTypeMindAndBody":                   "身心运动",
	"HKWorkoutActivityTypeBarre":                         "芭杆",
	"HKWorkoutActivityTypePilates":                       "普拉提",
}

func HealthCacheEnabled() bool {
	return dal.Redis != nil
}

func DailyHealthCacheKey(date string) string {
	prefix := strings.Trim(config.Cfg.Redis.KeyPrefix, ":")
	if prefix == "" {
		prefix = config.ServerName
	}
	return fmt.Sprintf("%s:health:daily:%s", prefix, date)
}

func WriteDailyHealthCacheRange(ctx context.Context, start, end time.Time) (*HealthCacheWriteSummary, error) {
	if !HealthCacheEnabled() {
		return nil, ErrHealthCacheDisabled
	}

	startDay, endDay, err := normalizeDayRange(start, end)
	if err != nil {
		return nil, err
	}

	summary := &HealthCacheWriteSummary{
		StartDate: startDay.Format("2006-01-02"),
		EndDate:   endDay.Format("2006-01-02"),
	}

	for day := startDay; !day.After(endDay); day = day.AddDate(0, 0, 1) {
		cache, err := BuildDailyHealthCache(ctx, day)
		if err != nil {
			return nil, err
		}

		if cache.IsZero() {
			if err := dal.Redis.Del(ctx, DailyHealthCacheKey(cache.Date)).Err(); err != nil {
				return nil, err
			}
			summary.SkippedDates = append(summary.SkippedDates, cache.Date)
			continue
		}

		payload, err := json.Marshal(cache)
		if err != nil {
			return nil, err
		}

		if err := dal.Redis.Set(ctx, DailyHealthCacheKey(cache.Date), payload, 0).Err(); err != nil {
			return nil, err
		}

		summary.WrittenDates = append(summary.WrittenDates, cache.Date)
	}
	summary.WrittenCount = len(summary.WrittenDates)
	summary.SkippedCount = len(summary.SkippedDates)
	return summary, nil
}

func GetDailyHealthCacheRange(ctx context.Context, start, end time.Time) (*HealthCacheRange, error) {
	if !HealthCacheEnabled() {
		return nil, ErrHealthCacheDisabled
	}

	startDay, endDay, err := normalizeDayRange(start, end)
	if err != nil {
		return nil, err
	}

	result := &HealthCacheRange{
		StartDate:      startDay.Format("2006-01-02"),
		EndDate:        endDay.Format("2006-01-02"),
		StepCounts:     make([]DailyStepCount, 0),
		WorkoutRecords: make([]DailyWorkoutRecord, 0),
	}

	for day := startDay; !day.After(endDay); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		payload, err := dal.Redis.Get(ctx, DailyHealthCacheKey(date)).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			return nil, err
		}

		item, err := decodeDailyHealthCache(payload)
		if err != nil {
			return nil, fmt.Errorf("解析redis缓存失败 date=%s: %w", date, err)
		}
		appendDailyHealthCache(result, item)
	}

	return result, nil
}

func BuildDailyHealthCache(ctx context.Context, day time.Time) (*DailyHealthCache, error) {
	dayStart, dayEnd, date := dayBounds(day)

	stepCount, err := queryDailyStepCount(ctx, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	workoutRecords, err := queryDailyWorkoutRecords(ctx, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	return &DailyHealthCache{
		Date:           date,
		StepCount:      stepCount,
		WorkoutRecords: workoutRecords,
	}, nil
}

func queryDailyStepCount(ctx context.Context, dayStart, dayEnd time.Time) (int64, error) {
	var row struct {
		StepCount float64
	}

	err := dal.DB.WithContext(ctx).Raw(`
WITH daily AS (
	SELECT
		COALESCE(SUM(CASE WHEN COALESCE(source_name, '') LIKE '%Watch%' THEN value ELSE 0 END), 0) AS watch_steps,
		COALESCE(SUM(CASE WHEN COALESCE(source_name, '') LIKE '%iPhone%' THEN value ELSE 0 END), 0) AS iphone_steps,
		COALESCE(SUM(CASE WHEN COALESCE(source_name, '') LIKE '%小米%' THEN value ELSE 0 END), 0) AS xiaomi_steps
	FROM health_records
	WHERE type = 'StepCount'
		AND start_date >= ?
		AND start_date < ?
)
SELECT
	CASE
		WHEN watch_steps > 0 THEN watch_steps
		WHEN iphone_steps > 0 THEN iphone_steps
		ELSE xiaomi_steps
	END AS step_count
FROM daily`, dayStart, dayEnd).Scan(&row).Error
	if err != nil {
		return 0, err
	}

	return roundSteps(row.StepCount), nil
}

func queryDailyWorkoutRecords(ctx context.Context, dayStart, dayEnd time.Time) ([]DailyWorkoutCacheRecord, error) {
	var rows []workoutAggregateRow
	err := dal.DB.WithContext(ctx).Raw(`
SELECT
	w.start_date,
	w.end_date,
	w.workout_activity_type,
	ROUND(COALESCE(
		NULLIF(w.duration, 0)::numeric,
		(EXTRACT(EPOCH FROM (w.end_date - w.start_date)) / 60)::numeric,
		0::numeric
	), 1)::float8 AS duration_minutes,
	ROUND(COALESCE(
		NULLIF(w.total_distance, 0)::numeric,
		(
			SELECT SUM(hr.value)::numeric
			FROM health_records hr
			WHERE hr.type IN ('DistanceWalkingRunning', 'DistanceCycling', 'DistanceSwimming')
				AND hr.start_date < w.end_date
				AND hr.end_date > w.start_date
		),
		0::numeric
	), 2)::float8 AS distance_km,
	ROUND(COALESCE(
		NULLIF(w.total_energy_burned, 0)::numeric,
		(
			SELECT SUM(hr.value)::numeric
			FROM health_records hr
			WHERE hr.type = 'ActiveEnergyBurned'
				AND hr.start_date < w.end_date
				AND hr.end_date > w.start_date
		),
		0::numeric
	), 1)::float8 AS energy_kcal,
	w.source_name
FROM workouts w
WHERE w.start_date >= ?
	AND w.start_date < ?
ORDER BY w.start_date DESC`, dayStart, dayEnd).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	loc := healthCacheLocation()
	records := make([]DailyWorkoutCacheRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, DailyWorkoutCacheRecord{
			StartTime:       row.StartDate.In(loc),
			EndTime:         row.EndDate.In(loc),
			WorkoutType:     row.WorkoutActivityType,
			WorkoutName:     workoutActivityName(row.WorkoutActivityType),
			DurationMinutes: roundFloat(row.DurationMinutes, 1),
			DistanceKM:      roundFloat(row.DistanceKM, 2),
			EnergyKCal:      roundFloat(row.EnergyKCal, 1),
			SourceName:      row.SourceName,
		})
	}

	return records, nil
}

func decodeDailyHealthCache(payload string) (DailyHealthCache, error) {
	if strings.Contains(payload, `"step_count"`) || strings.Contains(payload, `"workout_records"`) {
		var item DailyHealthCache
		if err := json.Unmarshal([]byte(payload), &item); err != nil {
			return DailyHealthCache{}, err
		}
		return item, nil
	}

	// Backward compatibility for values written before Redis storage was split from API formatting.
	var legacy struct {
		Date      string `json:"date"`
		StepStats struct {
			Steps int64 `json:"steps"`
		} `json:"step_stats"`
		Workouts []struct {
			StartTime       time.Time `json:"start_time"`
			EndTime         time.Time `json:"end_time"`
			ActivityType    string    `json:"activity_type"`
			ActivityName    string    `json:"activity_name"`
			DurationMinutes float64   `json:"duration_minutes"`
			DistanceKM      float64   `json:"distance_km"`
			EnergyKCal      float64   `json:"energy_kcal"`
			SourceName      string    `json:"source_name"`
		} `json:"workouts"`
	}
	if err := json.Unmarshal([]byte(payload), &legacy); err != nil {
		return DailyHealthCache{}, err
	}

	item := DailyHealthCache{
		Date:      legacy.Date,
		StepCount: legacy.StepStats.Steps,
	}
	for _, workout := range legacy.Workouts {
		item.WorkoutRecords = append(item.WorkoutRecords, DailyWorkoutCacheRecord{
			StartTime:       workout.StartTime,
			EndTime:         workout.EndTime,
			WorkoutType:     workout.ActivityType,
			WorkoutName:     workout.ActivityName,
			DurationMinutes: workout.DurationMinutes,
			DistanceKM:      workout.DistanceKM,
			EnergyKCal:      workout.EnergyKCal,
			SourceName:      workout.SourceName,
		})
	}
	return item, nil
}

func appendDailyHealthCache(result *HealthCacheRange, item DailyHealthCache) {
	if item.StepCount > 0 {
		result.StepCounts = append(result.StepCounts, DailyStepCount{
			Date:      item.Date,
			StepCount: item.StepCount,
		})
	}

	for _, record := range item.WorkoutRecords {
		duration := roundFloat(record.DurationMinutes, 1)
		distance := roundFloat(record.DistanceKM, 2)
		energy := roundFloat(record.EnergyKCal, 1)

		result.WorkoutRecords = append(result.WorkoutRecords, DailyWorkoutRecord{
			Date:            item.Date,
			StartTime:       record.StartTime,
			EndTime:         record.EndTime,
			TimeRange:       formatTimeRange(record.StartTime, record.EndTime),
			WorkoutType:     record.WorkoutType,
			WorkoutName:     record.WorkoutName,
			DurationMinutes: duration,
			DurationText:    formatMinutes(duration),
			DistanceKM:      distance,
			DistanceText:    formatDistance(distance),
			EnergyKCal:      energy,
			EnergyText:      formatEnergy(energy),
			SourceName:      record.SourceName,
		})
	}
}

func healthCacheLocation() *time.Location {
	loc, err := time.LoadLocation(config.Cfg.Server.Zone)
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

func normalizeDayRange(start, end time.Time) (time.Time, time.Time, error) {
	startDay, _, _ := dayBounds(start)
	endDay, _, _ := dayBounds(end)
	if startDay.After(endDay) {
		return time.Time{}, time.Time{}, errors.New("start_date不能晚于end_date")
	}
	return startDay, endDay, nil
}

func dayBounds(day time.Time) (time.Time, time.Time, string) {
	loc := healthCacheLocation()
	localDay := day.In(loc)
	start := time.Date(localDay.Year(), localDay.Month(), localDay.Day(), 0, 0, 0, 0, loc)
	return start, start.AddDate(0, 0, 1), start.Format("2006-01-02")
}

func roundSteps(value float64) int64 {
	return int64(math.Round(value))
}

func roundFloat(value float64, places int) float64 {
	pow := math.Pow10(places)
	return math.Round(value*pow) / pow
}

func workoutActivityName(activityType string) string {
	if name, ok := workoutActivityDisplayNames[activityType]; ok {
		return name
	}

	name := strings.TrimPrefix(activityType, "HKWorkoutActivityType")
	if name == "" {
		return activityType
	}
	return name
}

func formatMinutes(value float64) string {
	return fmt.Sprintf("%.1f 分钟", value)
}

func formatDistance(value float64) string {
	return fmt.Sprintf("%.2f km", value)
}

func formatEnergy(value float64) string {
	return fmt.Sprintf("%.1f kcal", value)
}

func formatTimeRange(start, end time.Time) string {
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		return fmt.Sprintf("%s - %s", start.Format("15:04"), end.Format("15:04"))
	}
	return fmt.Sprintf("%s - %s", start.Format("01-02 15:04"), end.Format("01-02 15:04"))
}

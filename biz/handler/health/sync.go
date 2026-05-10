package health

import (
	"apple_health/biz/dal"
	"apple_health/biz/model"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var hkQuantityTypeRawValueMap = map[int]string{
	1:  "HKQuantityTypeIdentifierStepCount",
	2:  "HKQuantityTypeIdentifierDistanceWalkingRunning",
	3:  "HKQuantityTypeIdentifierDistanceCycling",
	4:  "HKQuantityTypeIdentifierDistanceWheelchair",
	5:  "HKQuantityTypeIdentifierBasalEnergyBurned",
	6:  "HKQuantityTypeIdentifierActiveEnergyBurned",
	7:  "HKQuantityTypeIdentifierFlightsClimbed",
	8:  "HKQuantityTypeIdentifierNikeFuel",
	9:  "HKQuantityTypeIdentifierAppleExerciseTime",
	10: "HKQuantityTypeIdentifierPushCount",
	11: "HKQuantityTypeIdentifierDistanceSwimming",
	12: "HKQuantityTypeIdentifierSwimmingStrokeCount",
	13: "HKQuantityTypeIdentifierVO2Max",
	14: "HKQuantityTypeIdentifierHeartRate",
	15: "HKQuantityTypeIdentifierRestingHeartRate",
	16: "HKQuantityTypeIdentifierWalkingHeartRateAverage",
	17: "HKQuantityTypeIdentifierHeartRateVariabilitySDNN",
	18: "HKQuantityTypeIdentifierStepCount",
	19: "HKQuantityTypeIdentifierDistanceWalkingRunning",
	20: "HKQuantityTypeIdentifierHeight",
	21: "HKQuantityTypeIdentifierBodyMass",
	22: "HKQuantityTypeIdentifierBodyMassIndex",
	23: "HKQuantityTypeIdentifierLeanBodyMass",
	24: "HKQuantityTypeIdentifierBodyFatPercentage",
	25: "HKQuantityTypeIdentifierWaistCircumference",
	26: "HKQuantityTypeIdentifierActiveEnergyBurned",
	27: "HKQuantityTypeIdentifierBasalEnergyBurned",
	28: "HKQuantityTypeIdentifierFlightsClimbed",
	29: "HKQuantityTypeIdentifierStepCount",
	30: "HKQuantityTypeIdentifierDistanceWalkingRunning",
	31: "HKQuantityTypeIdentifierDistanceCycling",
	32: "HKQuantityTypeIdentifierDistanceSwimming",
	33: "HKQuantityTypeIdentifierAppleExerciseTime",
	34: "HKQuantityTypeIdentifierAppleStandTime",
	35: "HKQuantityTypeIdentifierWalkingSpeed",
	36: "HKQuantityTypeIdentifierWalkingStepLength",
	37: "HKQuantityTypeIdentifierWalkingAsymmetryPercentage",
	38: "HKQuantityTypeIdentifierWalkingDoubleSupportPercentage",
	39: "HKQuantityTypeIdentifierSixMinuteWalkTestDistance",
	40: "HKQuantityTypeIdentifierEnvironmentalSoundReduction",
	41: "HKQuantityTypeIdentifierEnvironmentalSoundExposure",
	42: "HKQuantityTypeIdentifierHeadphoneAudioExposure",
	43: "HKQuantityTypeIdentifierNumberOfTimesFallen",
	44: "HKQuantityTypeIdentifierSleepAnalysis",
	45: "HKQuantityTypeIdentifierHeartRate",
	46: "HKQuantityTypeIdentifierRestingHeartRate",
	47: "HKQuantityTypeIdentifierWalkingHeartRateAverage",
	48: "HKQuantityTypeIdentifierHeartRateVariabilitySDNN",
	49: "HKQuantityTypeIdentifierOxygenSaturation",
	50: "HKQuantityTypeIdentifierPeripheralPerfusionIndex",
	51: "HKQuantityTypeIdentifierBloodPressureSystolic",
	52: "HKQuantityTypeIdentifierBloodPressureDiastolic",
	53: "HKQuantityTypeIdentifierRespiratoryRate",
	54: "HKQuantityTypeIdentifierBodyTemperature",
	55: "HKQuantityTypeIdentifierElectrodermalActivity",
	56: "HKQuantityTypeIdentifierInhalerUsage",
	57: "HKQuantityTypeIdentifierInsulinDelivery",
	58: "HKQuantityTypeIdentifierBloodGlucose",
	59: "HKQuantityTypeIdentifierNumberOfAlcoholicBeverages",
	60: "HKQuantityTypeIdentifierBloodAlcoholContent",
	61: "HKQuantityTypeIdentifierCervicalMucusQuality",
	62: "HKQuantityTypeIdentifierOvulationTestResult",
	63: "HKQuantityTypeIdentifierMenstrualFlow",
	64: "HKQuantityTypeIdentifierIntermenstrualBleeding",
	65: "HKQuantityTypeIdentifierSexualActivity",
	66: "HKQuantityTypeIdentifierMindfulSession",
	67: "HKQuantityTypeIdentifierAppleMoveTime",
	68: "HKQuantityTypeIdentifierUnderwaterDepth",
	69: "HKQuantityTypeIdentifierUnderwaterTemperature",
	70: "HKQuantityTypeIdentifierWaterTemperature",
	71: "HKQuantityTypeIdentifierSurfingSportsMaximumSpeed",
	72: "HKQuantityTypeIdentifierDownhillSkiingSpeed",
	73: "HKQuantityTypeIdentifierSnowboardingSpeed",
	74: "HKQuantityTypeIdentifierCrossCountrySkiingSpeed",
	75: "HKQuantityTypeIdentifierRunningSpeed",
	76: "HKQuantityTypeIdentifierRunningStrideLength",
	77: "HKQuantityTypeIdentifierRunningVerticalOscillation",
	78: "HKQuantityTypeIdentifierRunningGroundContactTime",
	79: "HKQuantityTypeIdentifierRunningPower",
	80: "HKQuantityTypeIdentifierCyclingSpeed",
	81: "HKQuantityTypeIdentifierCyclingPower",
	82: "HKQuantityTypeIdentifierCyclingCadence",
	83: "HKQuantityTypeIdentifierCyclingFunctionalThresholdPower",
	84: "HKQuantityTypeIdentifierTimeInDaylight",
	85: "HKQuantityTypeIdentifierSleepDuration",
	86: "HKQuantityTypeIdentifierTimeAsleep",
	87: "HKQuantityTypeIdentifierSleepRem",
	88: "HKQuantityTypeIdentifierSleepDeep",
	89: "HKQuantityTypeIdentifierSleepCore",
	90: "HKQuantityTypeIdentifierSleepAwake",
}

var hkQuantityTypeRawValueRegex = regexp.MustCompile(`HKQuantityTypeIdentifier\(rawValue:\s*(\d+)\)`)
var hkCategoryTypeRawValueRegex = regexp.MustCompile(`HKCategoryTypeIdentifier\(rawValue:\s*(\d+)\)`)

func normalizeHealthRecordType(t string) (typeName string, hkQuantityType string) {
	matches := hkQuantityTypeRawValueRegex.FindStringSubmatch(t)
	if len(matches) == 2 {
		rawVal, err := strconv.Atoi(matches[1])
		if err == nil {
			if name, ok := hkQuantityTypeRawValueMap[rawVal]; ok {
				hkQuantityType = name
				typeName = strings.TrimPrefix(name, "HKQuantityTypeIdentifier")
				return
			}
		}
		hkQuantityType = t
		typeName = t
		return
	}

	matches = hkCategoryTypeRawValueRegex.FindStringSubmatch(t)
	if len(matches) == 2 {
		hkQuantityType = t
		typeName = strings.TrimPrefix(t, "HKCategoryTypeIdentifier")
		return
	}

	if strings.HasPrefix(t, "HKQuantityTypeIdentifier") {
		hkQuantityType = t
		typeName = strings.TrimPrefix(t, "HKQuantityTypeIdentifier")
		return
	}

	if strings.HasPrefix(t, "HKCategoryTypeIdentifier") {
		hkQuantityType = t
		typeName = strings.TrimPrefix(t, "HKCategoryTypeIdentifier")
		return
	}

	hkQuantityType = t
	typeName = t
	return
}

var workoutRawValueMap = map[int]string{
	1:  "HKWorkoutActivityTypeAmericanFootball",
	2:  "HKWorkoutActivityTypeArchery",
	3:  "HKWorkoutActivityTypeAustralianFootball",
	4:  "HKWorkoutActivityTypeBadminton",
	5:  "HKWorkoutActivityTypeBaseball",
	6:  "HKWorkoutActivityTypeBasketball",
	7:  "HKWorkoutActivityTypeBowling",
	9:  "HKWorkoutActivityTypeClimbing",
	10: "HKWorkoutActivityTypeCricket",
	11: "HKWorkoutActivityTypeCrossTraining",
	12: "HKWorkoutActivityTypeCurling",
	13: "HKWorkoutActivityTypeCycling",
	14: "HKWorkoutActivityTypeDance",
	16: "HKWorkoutActivityTypeElliptical",
	17: "HKWorkoutActivityTypeEquestrianSports",
	18: "HKWorkoutActivityTypeFencing",
	19: "HKWorkoutActivityTypeFishing",
	20: "HKWorkoutActivityTypeFunctionalStrengthTraining",
	21: "HKWorkoutActivityTypeGolf",
	22: "HKWorkoutActivityTypeGymnastics",
	23: "HKWorkoutActivityTypeHandball",
	24: "HKWorkoutActivityTypeHiking",
	25: "HKWorkoutActivityTypeHockey",
	26: "HKWorkoutActivityTypeHunting",
	27: "HKWorkoutActivityTypeLacrosse",
	28: "HKWorkoutActivityTypeMartialArts",
	29: "HKWorkoutActivityTypeMindAndBody",
	31: "HKWorkoutActivityTypePaddleSports",
	32: "HKWorkoutActivityTypePlay",
	33: "HKWorkoutActivityTypePreparationAndRecovery",
	34: "HKWorkoutActivityTypeRacquetball",
	35: "HKWorkoutActivityTypeRowing",
	36: "HKWorkoutActivityTypeRugby",
	37: "HKWorkoutActivityTypeRunning",
	38: "HKWorkoutActivityTypeSailing",
	39: "HKWorkoutActivityTypeSkatingSports",
	40: "HKWorkoutActivityTypeSnowSports",
	41: "HKWorkoutActivityTypeSoccer",
	42: "HKWorkoutActivityTypeSoftball",
	43: "HKWorkoutActivityTypeSquash",
	44: "HKWorkoutActivityTypeStairClimbing",
	45: "HKWorkoutActivityTypeSurfingSports",
	46: "HKWorkoutActivityTypeSwimming",
	47: "HKWorkoutActivityTypeTableTennis",
	48: "HKWorkoutActivityTypeTennis",
	49: "HKWorkoutActivityTypeTrackAndField",
	50: "HKWorkoutActivityTypeTraditionalStrengthTraining",
	51: "HKWorkoutActivityTypeVolleyball",
	52: "HKWorkoutActivityTypeWalking",
	53: "HKWorkoutActivityTypeWaterFitness",
	54: "HKWorkoutActivityTypeWaterPolo",
	55: "HKWorkoutActivityTypeWaterSports",
	56: "HKWorkoutActivityTypeWrestling",
	57: "HKWorkoutActivityTypeYoga",
	58: "HKWorkoutActivityTypeBarre",
	59: "HKWorkoutActivityTypeCoreTraining",
	60: "HKWorkoutActivityTypeCrossCountrySkiing",
	61: "HKWorkoutActivityTypeDownhillSkiing",
	62: "HKWorkoutActivityTypeFlexibility",
	63: "HKWorkoutActivityTypeHighIntensityIntervalTraining",
	64: "HKWorkoutActivityTypeJumpRope",
	65: "HKWorkoutActivityTypeKickboxing",
	66: "HKWorkoutActivityTypePilates",
	67: "HKWorkoutActivityTypeSnowboarding",
	68: "HKWorkoutActivityTypeStairs",
	69: "HKWorkoutActivityTypeStepTraining",
	70: "HKWorkoutActivityTypeTaiChi",
	71: "HKWorkoutActivityTypeMixedCardio",
	72: "HKWorkoutActivityTypeHandCycling",
	73: "HKWorkoutActivityTypeDiscSports",
	74: "HKWorkoutActivityTypeFitnessGaming",
	75: "HKWorkoutActivityTypeCardioDance",
	76: "HKWorkoutActivityTypeSocialDance",
	77: "HKWorkoutActivityTypePickleball",
	78: "HKWorkoutActivityTypeShuffleboard",
}

var rawValueRegex = regexp.MustCompile(`HKWorkoutActivityType\(rawValue:\s*(\d+)\)`)

func normalizeWorkoutActivityType(t string) string {
	matches := rawValueRegex.FindStringSubmatch(t)
	if len(matches) == 2 {
		rawVal, err := strconv.Atoi(matches[1])
		if err == nil {
			if name, ok := workoutRawValueMap[rawVal]; ok {
				return name
			}
			return fmt.Sprintf("HKWorkoutActivityTypeUnknown(%d)", rawVal)
		}
	}
	return t
}

type SyncHealthDataReq struct {
	StartDate         string                  `json:"start_date" binding:"required"`
	EndDate           string                  `json:"end_date" binding:"required"`
	HealthRecords     []model.HealthRecord    `json:"health_records,omitempty"`
	ActivitySummaries []model.ActivitySummary `json:"activity_summaries,omitempty"`
	SleepAnalyses     []model.SleepAnalysis   `json:"sleep_analyses,omitempty"`
	Workouts          []model.Workout         `json:"workouts,omitempty"`
	WorkoutRoutes     []model.WorkoutRoute    `json:"workout_routes,omitempty"`
	WorkoutLocations  []model.WorkoutLocation `json:"workout_locations,omitempty"`
}

type SyncHealthDataResp struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    *SyncHealthDataSummary `json:"data,omitempty"`
}

type SyncHealthDataSummary struct {
	HealthRecordCount    int `json:"health_record_count"`
	ActivitySummaryCount int `json:"activity_summary_count"`
	SleepAnalysisCount   int `json:"sleep_analysis_count"`
	WorkoutCount         int `json:"workout_count"`
	WorkoutRouteCount    int `json:"workout_route_count"`
	WorkoutLocationCount int `json:"workout_location_count"`
}

// SyncHealthData App上传指定时间段的健康数据
// @Summary App上传指定时间段的健康数据
// @Description App选定时间段后，将该时间段内的健康数据上传到服务端
// @Tags 健康数据
// @Accept application/json
// @Produce json
// @Param body body SyncHealthDataReq true "上传数据"
// @Success 200 {object} SyncHealthDataResp
// @Failure 400 {object} SyncHealthDataResp
// @Failure 500 {object} SyncHealthDataResp
// @Router /api/health/sync [post]
func SyncHealthData(c *gin.Context) {
	var req SyncHealthDataReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SyncHealthDataResp{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误，请提供 start_date 和 end_date",
		})
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, SyncHealthDataResp{
			Code:    http.StatusBadRequest,
			Message: "start_date 格式错误，请使用 RFC3339 格式（如 2024-01-01T00:00:00Z）",
		})
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, SyncHealthDataResp{
			Code:    http.StatusBadRequest,
			Message: "end_date 格式错误，请使用 RFC3339 格式（如 2024-12-31T23:59:59Z）",
		})
		return
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, SyncHealthDataResp{
			Code:    http.StatusBadRequest,
			Message: "start_date 不能晚于 end_date",
		})
		return
	}

	db := dal.DB
	summary := &SyncHealthDataSummary{}

	if len(req.HealthRecords) > 0 {
		for i := range req.HealthRecords {
			typeName, hkQuantityType := normalizeHealthRecordType(req.HealthRecords[i].Type)
			req.HealthRecords[i].Type = typeName
			if req.HealthRecords[i].HKQuantityType == "" {
				req.HealthRecords[i].HKQuantityType = hkQuantityType
			}
		}
		days := extractDays(req.HealthRecords, func(r model.HealthRecord) time.Time { return r.StartDate })
		deleteByDays(db, days, "health_records", "start_date")
		if err := db.Create(&req.HealthRecords).Error; err != nil {
			c.JSON(http.StatusInternalServerError, SyncHealthDataResp{
				Code:    http.StatusInternalServerError,
				Message: "健康记录写入失败: " + err.Error(),
			})
			return
		}
		summary.HealthRecordCount = len(req.HealthRecords)
	}

	if len(req.ActivitySummaries) > 0 {
		days := make(map[string]bool)
		for _, a := range req.ActivitySummaries {
			days[a.DateComponents] = true
		}
		for day := range days {
			db.Where("date_components = ?", day).Delete(&model.ActivitySummary{})
		}
		if err := db.Create(&req.ActivitySummaries).Error; err != nil {
			c.JSON(http.StatusInternalServerError, SyncHealthDataResp{
				Code:    http.StatusInternalServerError,
				Message: "活动摘要写入失败: " + err.Error(),
			})
			return
		}
		summary.ActivitySummaryCount = len(req.ActivitySummaries)
	}

	if len(req.SleepAnalyses) > 0 {
		days := extractDays(req.SleepAnalyses, func(r model.SleepAnalysis) time.Time { return r.StartDate })
		deleteByDays(db, days, "sleep_analyses", "start_date")
		if err := db.Create(&req.SleepAnalyses).Error; err != nil {
			c.JSON(http.StatusInternalServerError, SyncHealthDataResp{
				Code:    http.StatusInternalServerError,
				Message: "睡眠分析写入失败: " + err.Error(),
			})
			return
		}
		summary.SleepAnalysisCount = len(req.SleepAnalyses)
	}

	if len(req.Workouts) > 0 {
		for i := range req.Workouts {
			req.Workouts[i].WorkoutActivityType = normalizeWorkoutActivityType(req.Workouts[i].WorkoutActivityType)
		}
		days := extractDays(req.Workouts, func(r model.Workout) time.Time { return r.StartDate })
		deleteByDays(db, days, "workouts", "start_date")
		if err := db.Create(&req.Workouts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, SyncHealthDataResp{
				Code:    http.StatusInternalServerError,
				Message: "锻炼记录写入失败: " + err.Error(),
			})
			return
		}
		summary.WorkoutCount = len(req.Workouts)
	}

	if len(req.WorkoutRoutes) > 0 {
		days := extractDays(req.WorkoutRoutes, func(r model.WorkoutRoute) time.Time { return r.StartDate })
		deleteByDays(db, days, "workout_routes", "start_date")
		if err := db.Create(&req.WorkoutRoutes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, SyncHealthDataResp{
				Code:    http.StatusInternalServerError,
				Message: "锻炼路线写入失败: " + err.Error(),
			})
			return
		}
		summary.WorkoutRouteCount = len(req.WorkoutRoutes)
	}

	if len(req.WorkoutLocations) > 0 {
		days := extractDays(req.WorkoutLocations, func(r model.WorkoutLocation) time.Time { return r.Timestamp })
		deleteByDays(db, days, "workout_locations", "timestamp")
		if err := db.Create(&req.WorkoutLocations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, SyncHealthDataResp{
				Code:    http.StatusInternalServerError,
				Message: "锻炼位置写入失败: " + err.Error(),
			})
			return
		}
		summary.WorkoutLocationCount = len(req.WorkoutLocations)
	}

	c.JSON(http.StatusOK, SyncHealthDataResp{
		Code:    http.StatusOK,
		Message: "同步成功",
		Data:    summary,
	})
}

var cstLoc = time.FixedZone("CST", 8*3600)

func extractDays[T any](records []T, getTime func(T) time.Time) map[string]bool {
	days := make(map[string]bool)
	for _, r := range records {
		t := getTime(r).In(cstLoc)
		days[t.Format("2006-01-02")] = true
	}
	return days
}

func deleteByDays(db *gorm.DB, days map[string]bool, table string, column string) {
	for day := range days {
		dayStart, _ := time.ParseInLocation("2006-01-02", day, cstLoc)
		dayEnd := dayStart.Add(24 * time.Hour)
		db.Table(table).Where(column+" >= ? AND "+column+" < ?", dayStart, dayEnd).Delete(nil)
	}
}

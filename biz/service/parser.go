package service

import (
	"apple_health/biz/model"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type HealthData struct {
	XMLName           xml.Name             `xml:"HealthData"`
	ExportDate        string               `xml:"exportDate,attr"`
	Records           []Record             `xml:"Record"`
	Workouts          []WorkoutXML         `xml:"Workout"`
	WorkoutRoutes     []WorkoutRouteXML    `xml:"WorkoutRoute"`
	ActivitySummaries []ActivitySummaryXML `xml:"ActivitySummary"`
}

type Record struct {
	Type          string          `xml:"type,attr"`
	SourceName    string          `xml:"sourceName,attr"`
	SourceVersion string          `xml:"sourceVersion,attr"`
	Unit          string          `xml:"unit,attr"`
	Value         string          `xml:"value,attr"`
	StartDate     string          `xml:"startDate,attr"`
	EndDate       string          `xml:"endDate,attr"`
	CreationDate  string          `xml:"creationDate,attr"`
	Device        string          `xml:"device,attr"`
	Metadata      []MetadataEntry `xml:"MetadataEntry"`
}

type WorkoutXML struct {
	WorkoutActivityType   string            `xml:"workoutActivityType,attr"`
	Duration              string            `xml:"duration,attr"`
	DurationUnit          string            `xml:"durationUnit,attr"`
	TotalDistance         string            `xml:"totalDistance,attr"`
	TotalDistanceUnit     string            `xml:"totalDistanceUnit,attr"`
	TotalEnergyBurned     string            `xml:"totalEnergyBurned,attr"`
	TotalEnergyBurnedUnit string            `xml:"totalEnergyBurnedUnit,attr"`
	SourceName            string            `xml:"sourceName,attr"`
	SourceVersion         string            `xml:"sourceVersion,attr"`
	StartDate             string            `xml:"startDate,attr"`
	EndDate               string            `xml:"endDate,attr"`
	CreationDate          string            `xml:"creationDate,attr"`
	Device                string            `xml:"device,attr"`
	Metadata              []MetadataEntry   `xml:"MetadataEntry"`
	WorkoutRoutes         []WorkoutRouteXML `xml:"WorkoutRoute"`
}

type ActivitySummaryXML struct {
	DateComponents         string `xml:"dateComponents,attr"`
	ActiveEnergyBurned     string `xml:"activeEnergyBurned,attr"`
	ActiveEnergyBurnedGoal string `xml:"activeEnergyBurnedGoal,attr"`
	ActiveEnergyBurnedUnit string `xml:"activeEnergyBurnedUnit,attr"`
	AppleExerciseTime      string `xml:"appleExerciseTime,attr"`
	AppleExerciseTimeGoal  string `xml:"appleExerciseTimeGoal,attr"`
	AppleExerciseTimeUnit  string `xml:"appleExerciseTimeUnit,attr"`
	AppleStandTime         string `xml:"appleStandHours,attr"`
	AppleStandTimeGoal     string `xml:"appleStandHoursGoal,attr"`
	AppleStandTimeUnit     string `xml:"appleStandHoursUnit,attr"`
}

type MetadataEntry struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type WorkoutRouteXML struct {
	SourceName    string           `xml:"sourceName,attr"`
	StartDate     string           `xml:"startDate,attr"`
	EndDate       string           `xml:"endDate,attr"`
	Locations     []LocationXML    `xml:"Location"`
	FileReference FileReferenceXML `xml:"FileReference"`
}

type FileReferenceXML struct {
	Path string `xml:"path,attr"`
}

type LocationXML struct {
	Latitude           string `xml:"latitude,attr"`
	Longitude          string `xml:"longitude,attr"`
	Altitude           string `xml:"altitude,attr"`
	Timestamp          string `xml:"timestamp,attr"`
	HorizontalAccuracy string `xml:"horizontalAccuracy,attr"`
	VerticalAccuracy   string `xml:"verticalAccuracy,attr"`
	Speed              string `xml:"speed,attr"`
	Course             string `xml:"course,attr"`
}

type GPX struct {
	XMLName xml.Name `xml:"gpx"`
	Trk     Trk      `xml:"trk"`
}

type Trk struct {
	Trkseg []Trkseg `xml:"trkseg"`
}

type Trkseg struct {
	Trkpt []Trkpt `xml:"trkpt"`
}

type Trkpt struct {
	Lat        string      `xml:"lat,attr"`
	Lon        string      `xml:"lon,attr"`
	Ele        string      `xml:"ele"`
	Time       string      `xml:"time"`
	Extensions *Extensions `xml:"extensions,omitempty"`
}

type Extensions struct {
	Speed  string `xml:"speed"`
	Course string `xml:"course"`
}

func ParseHealthData(reader io.Reader) (*HealthData, error) {
	var healthData HealthData
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&healthData); err != nil {
		return nil, fmt.Errorf("解析 XML 失败: %w", err)
	}
	return &healthData, nil
}

func parseAppleDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05 -0700",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05 -0700 MST",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期: %s", dateStr)
}

func ConvertRecord(record Record) (*model.HealthRecord, error) {
	startDate, err := parseAppleDate(record.StartDate)
	if err != nil {
		return nil, err
	}

	endDate, err := parseAppleDate(record.EndDate)
	if err != nil {
		return nil, err
	}

	creationDate, err := parseAppleDate(record.CreationDate)
	if err != nil {
		return nil, err
	}

	var value float64
	fmt.Sscanf(record.Value, "%f", &value)

	hkQuantityType := ""
	if strings.HasPrefix(record.Type, "HKQuantityTypeIdentifier") {
		hkQuantityType = strings.TrimPrefix(record.Type, "HKQuantityTypeIdentifier")
	} else if strings.HasPrefix(record.Type, "HKCategoryTypeIdentifier") {
		hkQuantityType = strings.TrimPrefix(record.Type, "HKCategoryTypeIdentifier")
	}

	metadataJSON := ""
	if len(record.Metadata) > 0 {
		var metadataPairs []string
		for _, m := range record.Metadata {
			metadataPairs = append(metadataPairs, fmt.Sprintf(`"%s":"%s"`, m.Key, m.Value))
		}
		metadataJSON = "{" + strings.Join(metadataPairs, ",") + "}"
	}

	return &model.HealthRecord{
		Type:           hkQuantityType,
		SourceName:     record.SourceName,
		SourceVersion:  record.SourceVersion,
		Unit:           record.Unit,
		Value:          value,
		StartDate:      startDate,
		EndDate:        endDate,
		CreationDate:   creationDate,
		Device:         record.Device,
		HKQuantityType: record.Type,
		Metadata:       metadataJSON,
	}, nil
}

func ConvertWorkout(workout WorkoutXML) (*model.Workout, error) {
	startDate, err := parseAppleDate(workout.StartDate)
	if err != nil {
		return nil, err
	}

	endDate, err := parseAppleDate(workout.EndDate)
	if err != nil {
		return nil, err
	}

	creationDate, err := parseAppleDate(workout.CreationDate)
	if err != nil {
		return nil, err
	}

	var duration, totalDistance, totalEnergyBurned float64
	fmt.Sscanf(workout.Duration, "%f", &duration)
	fmt.Sscanf(workout.TotalDistance, "%f", &totalDistance)
	fmt.Sscanf(workout.TotalEnergyBurned, "%f", &totalEnergyBurned)

	metadataJSON := ""
	if len(workout.Metadata) > 0 {
		var metadataPairs []string
		for _, m := range workout.Metadata {
			metadataPairs = append(metadataPairs, fmt.Sprintf(`"%s":"%s"`, m.Key, m.Value))
		}
		metadataJSON = "{" + strings.Join(metadataPairs, ",") + "}"
	}

	return &model.Workout{
		WorkoutActivityType:   workout.WorkoutActivityType,
		Duration:              duration,
		DurationUnit:          workout.DurationUnit,
		TotalDistance:         totalDistance,
		TotalDistanceUnit:     workout.TotalDistanceUnit,
		TotalEnergyBurned:     totalEnergyBurned,
		TotalEnergyBurnedUnit: workout.TotalEnergyBurnedUnit,
		SourceName:            workout.SourceName,
		SourceVersion:         workout.SourceVersion,
		StartDate:             startDate,
		EndDate:               endDate,
		CreationDate:          creationDate,
		Device:                workout.Device,
		Metadata:              metadataJSON,
	}, nil
}

func ConvertActivitySummary(summary ActivitySummaryXML) (*model.ActivitySummary, error) {
	var activeEnergyBurned, activeEnergyBurnedGoal float64
	var appleExerciseTime, appleExerciseTimeGoal float64
	var appleStandTime, appleStandTimeGoal float64

	fmt.Sscanf(summary.ActiveEnergyBurned, "%f", &activeEnergyBurned)
	fmt.Sscanf(summary.ActiveEnergyBurnedGoal, "%f", &activeEnergyBurnedGoal)
	fmt.Sscanf(summary.AppleExerciseTime, "%f", &appleExerciseTime)
	fmt.Sscanf(summary.AppleExerciseTimeGoal, "%f", &appleExerciseTimeGoal)
	fmt.Sscanf(summary.AppleStandTime, "%f", &appleStandTime)
	fmt.Sscanf(summary.AppleStandTimeGoal, "%f", &appleStandTimeGoal)

	return &model.ActivitySummary{
		DateComponents:         summary.DateComponents,
		ActiveEnergyBurned:     activeEnergyBurned,
		ActiveEnergyBurnedGoal: activeEnergyBurnedGoal,
		ActiveEnergyBurnedUnit: summary.ActiveEnergyBurnedUnit,
		AppleExerciseTime:      appleExerciseTime,
		AppleExerciseTimeGoal:  appleExerciseTimeGoal,
		AppleExerciseTimeUnit:  summary.AppleExerciseTimeUnit,
		AppleStandTime:         appleStandTime,
		AppleStandTimeGoal:     appleStandTimeGoal,
		AppleStandTimeUnit:     summary.AppleStandTimeUnit,
	}, nil
}

func ConvertWorkoutRoute(route WorkoutRouteXML) (*model.WorkoutRoute, []model.WorkoutLocation, error) {
	startDate, err := parseAppleDate(route.StartDate)
	if err != nil {
		return nil, nil, err
	}

	endDate, err := parseAppleDate(route.EndDate)
	if err != nil {
		return nil, nil, err
	}

	workoutRoute := &model.WorkoutRoute{
		SourceName: route.SourceName,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	var locations []model.WorkoutLocation
	for _, loc := range route.Locations {
		var lat, lon, alt, horAcc, verAcc, speed, course float64
		fmt.Sscanf(loc.Latitude, "%f", &lat)
		fmt.Sscanf(loc.Longitude, "%f", &lon)
		fmt.Sscanf(loc.Altitude, "%f", &alt)
		fmt.Sscanf(loc.HorizontalAccuracy, "%f", &horAcc)
		fmt.Sscanf(loc.VerticalAccuracy, "%f", &verAcc)
		fmt.Sscanf(loc.Speed, "%f", &speed)
		fmt.Sscanf(loc.Course, "%f", &course)

		timestamp, err := parseAppleDate(loc.Timestamp)
		if err != nil {
			continue
		}

		locations = append(locations, model.WorkoutLocation{
			Latitude:           lat,
			Longitude:          lon,
			Altitude:           alt,
			Timestamp:          timestamp,
			HorizontalAccuracy: horAcc,
			VerticalAccuracy:   verAcc,
			Speed:              speed,
			Course:             course,
		})
	}

	return workoutRoute, locations, nil
}

func ConvertSleepAnalysis(record Record) (*model.SleepAnalysis, error) {
	startDate, err := parseAppleDate(record.StartDate)
	if err != nil {
		return nil, err
	}

	endDate, err := parseAppleDate(record.EndDate)
	if err != nil {
		return nil, err
	}

	creationDate, err := parseAppleDate(record.CreationDate)
	if err != nil {
		return nil, err
	}

	return &model.SleepAnalysis{
		SourceName:    record.SourceName,
		SourceVersion: record.SourceVersion,
		Value:         record.Value,
		StartDate:     startDate,
		EndDate:       endDate,
		CreationDate:  creationDate,
		Device:        record.Device,
	}, nil
}

func ParseGPXFile(filePath string) ([]model.WorkoutLocation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开 GPX 文件失败: %w", err)
	}
	defer file.Close()

	var gpx GPX
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&gpx); err != nil {
		return nil, fmt.Errorf("解析 GPX 文件失败: %w", err)
	}

	var locations []model.WorkoutLocation
	for _, trkseg := range gpx.Trk.Trkseg {
		for _, trkpt := range trkseg.Trkpt {
			var lat, lon, alt float64
			fmt.Sscanf(trkpt.Lat, "%f", &lat)
			fmt.Sscanf(trkpt.Lon, "%f", &lon)
			fmt.Sscanf(trkpt.Ele, "%f", &alt)

			var timestamp time.Time
			if trkpt.Time != "" {
				timestamp, _ = time.Parse(time.RFC3339, trkpt.Time)
			}

			var speed, course float64
			if trkpt.Extensions != nil {
				fmt.Sscanf(trkpt.Extensions.Speed, "%f", &speed)
				fmt.Sscanf(trkpt.Extensions.Course, "%f", &course)
			}

			locations = append(locations, model.WorkoutLocation{
				Latitude:  lat,
				Longitude: lon,
				Altitude:  alt,
				Timestamp: timestamp,
				Speed:     speed,
				Course:    course,
			})
		}
	}

	return locations, nil
}

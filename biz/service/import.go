package service

import (
	"apple_health/biz/dal"
	"apple_health/biz/model"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/slog"
	"gorm.io/gorm"
)

var cstLoc = time.FixedZone("CST", 8*3600)

type ImportService struct {
	db *gorm.DB
}

func NewImportService() *ImportService {
	return &ImportService{
		db: dal.DB,
	}
}

func (s *ImportService) ImportFromZip(zipPath string) (*model.ImportLog, error) {
	fileInfo, err := os.Stat(zipPath)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	importLog := &model.ImportLog{
		FileName: filepath.Base(zipPath),
		FileSize: fileInfo.Size(),
		Status:   "processing",
	}

	if err := s.db.Create(importLog).Error; err != nil {
		return nil, fmt.Errorf("创建导入日志失败: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "apple_health_*")
	if err != nil {
		s.updateImportLog(importLog, 0, "failed", fmt.Sprintf("创建临时目录失败: %v", err))
		return importLog, nil
	}
	defer os.RemoveAll(tempDir)

	if err := s.unzip(zipPath, tempDir); err != nil {
		s.updateImportLog(importLog, 0, "failed", fmt.Sprintf("解压文件失败: %v", err))
		return importLog, nil
	}

	var exportXMLPath string
	possiblePaths := []string{
		filepath.Join(tempDir, "apple_health_export", "export.xml"),
		filepath.Join(tempDir, "export.xml"),
		filepath.Join(tempDir, "导出.xml"),
		filepath.Join(tempDir, "apple_health_export", "导出.xml"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			exportXMLPath = path
			break
		}
	}

	if exportXMLPath == "" {
		s.updateImportLog(importLog, 0, "failed", "找不到 export.xml 或 导出.xml 文件")
		return importLog, nil
	}

	recordCount, err := s.importFromXML(exportXMLPath, tempDir)
	if err != nil {
		s.updateImportLog(importLog, 0, "failed", fmt.Sprintf("导入数据失败: %v", err))
		return importLog, nil
	}

	s.updateImportLog(importLog, recordCount, "success", "")
	return importLog, nil
}

func (s *ImportService) unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ImportService) importFromXML(xmlPath string, tempDir string) (int, error) {
	file, err := os.Open(xmlPath)
	if err != nil {
		return 0, fmt.Errorf("打开 XML 文件失败: %w", err)
	}
	defer file.Close()

	healthData, err := ParseHealthData(file)
	if err != nil {
		return 0, fmt.Errorf("解析 XML 失败: %w", err)
	}

	slog.Infof("解析结果: Records=%d, Workouts=%d, WorkoutRoutes=%d, ActivitySummaries=%d",
		len(healthData.Records), len(healthData.Workouts),
		len(healthData.WorkoutRoutes), len(healthData.ActivitySummaries))

	totalRecords := 0
	batchSize := 1000

	var records []model.HealthRecord
	var sleepRecords []model.SleepAnalysis

	for _, record := range healthData.Records {
		if strings.Contains(record.Type, "SleepAnalysis") {
			sleepAnalysis, err := ConvertSleepAnalysis(record)
			if err != nil {
				slog.Warnf("转换睡眠记录失败: %v, 跳过该记录", err)
				continue
			}
			sleepRecords = append(sleepRecords, *sleepAnalysis)
		} else {
			healthRecord, err := ConvertRecord(record)
			if err != nil {
				slog.Warnf("转换记录失败: %v, 跳过该记录", err)
				continue
			}
			records = append(records, *healthRecord)
		}
	}

	if len(records) > 0 {
		s.deleteByDaysAndInsert(records, func(r model.HealthRecord) string {
			return r.StartDate.In(cstLoc).Format("2006-01-02")
		}, func(tx *gorm.DB, day string) {
			dayStart, _ := time.ParseInLocation("2006-01-02", day, cstLoc)
			dayEnd := dayStart.Add(24 * time.Hour)
			tx.Where("start_date >= ? AND start_date < ?", dayStart, dayEnd).Delete(&model.HealthRecord{})
		})
		totalRecords += len(records)
		slog.Infof("插入 %d 条健康记录", len(records))
	}

	if len(sleepRecords) > 0 {
		s.deleteByDaysAndInsertSleep(sleepRecords)
		totalRecords += len(sleepRecords)
		slog.Infof("插入 %d 条睡眠记录", len(sleepRecords))
	}

	var workouts []model.Workout
	for _, workout := range healthData.Workouts {
		w, err := ConvertWorkout(workout)
		if err != nil {
			slog.Warnf("转换锻炼记录失败: %v, 跳过该记录", err)
			continue
		}
		workouts = append(workouts, *w)
	}

	if len(workouts) > 0 {
		s.deleteByDaysAndInsertWorkouts(workouts)
		totalRecords += len(workouts)
		slog.Infof("插入 %d 条锻炼记录", len(workouts))
	}

	slog.Infof("开始处理锻炼路线数据")
	routeCount := 0
	totalRoutes := 0
	for i, workoutXML := range healthData.Workouts {
		totalRoutes += len(workoutXML.WorkoutRoutes)
		for _, routeXML := range workoutXML.WorkoutRoutes {
			route, locations, err := ConvertWorkoutRoute(routeXML)
			if err != nil {
				slog.Warnf("转换路线失败: %v, 跳过该记录", err)
				continue
			}

			if routeXML.FileReference.Path != "" {
				gpxPath := filepath.Join(tempDir, "apple_health_export", routeXML.FileReference.Path)
				if _, err := os.Stat(gpxPath); err == nil {
					gpxLocations, err := ParseGPXFile(gpxPath)
					if err != nil {
						slog.Warnf("解析 GPX 文件失败: %v, 文件: %s", err, gpxPath)
					} else {
						locations = gpxLocations
						slog.Infof("从 GPX 文件读取到 %d 个位置点", len(locations))
					}
				} else {
					slog.Warnf("GPX 文件不存在: %s", gpxPath)
				}
			}

			if i < len(workouts) {
				route.WorkoutID = workouts[i].ID
				if err := s.db.Create(route).Error; err != nil {
					slog.Warnf("插入路线失败: %v", err)
					continue
				}

				for j := range locations {
					locations[j].RouteID = route.ID
				}

				if len(locations) > 0 {
					if err := s.db.CreateInBatches(locations, 100).Error; err != nil {
						slog.Warnf("插入位置点失败: %v", err)
					} else {
						totalRecords += len(locations)
						routeCount++
					}
				}
			}
		}
	}
	slog.Infof("总共找到 %d 条路线数据，成功插入 %d 条路线", totalRoutes, routeCount)

	var activitySummaries []model.ActivitySummary
	for _, summary := range healthData.ActivitySummaries {
		s, err := ConvertActivitySummary(summary)
		if err != nil {
			slog.Warnf("转换活动摘要失败: %v, 跳过该记录", err)
			continue
		}
		activitySummaries = append(activitySummaries, *s)
	}

	if len(activitySummaries) > 0 {
		days := make(map[string]bool)
		for _, a := range activitySummaries {
			days[a.DateComponents] = true
		}
		for day := range days {
			s.db.Where("date_components = ?", day).Delete(&model.ActivitySummary{})
		}
		if err := s.db.CreateInBatches(activitySummaries, batchSize).Error; err != nil {
			slog.Warnf("批量插入活动摘要失败: %v", err)
		} else {
			totalRecords += len(activitySummaries)
		}
	}

	return totalRecords, nil
}

func (s *ImportService) deleteByDaysAndInsert(records []model.HealthRecord, getDay func(model.HealthRecord) string, deleteFn func(*gorm.DB, string)) {
	days := make(map[string]bool)
	for _, r := range records {
		days[getDay(r)] = true
	}

	s.db.Transaction(func(tx *gorm.DB) error {
		for day := range days {
			deleteFn(tx, day)
		}
		return tx.CreateInBatches(records, 1000).Error
	})
}

func (s *ImportService) deleteByDaysAndInsertSleep(records []model.SleepAnalysis) {
	days := make(map[string]bool)
	for _, r := range records {
		days[r.StartDate.In(cstLoc).Format("2006-01-02")] = true
	}

	s.db.Transaction(func(tx *gorm.DB) error {
		for day := range days {
			dayStart, _ := time.ParseInLocation("2006-01-02", day, cstLoc)
			dayEnd := dayStart.Add(24 * time.Hour)
			tx.Where("start_date >= ? AND start_date < ?", dayStart, dayEnd).Delete(&model.SleepAnalysis{})
		}
		return tx.CreateInBatches(records, 1000).Error
	})
}

func (s *ImportService) deleteByDaysAndInsertWorkouts(workouts []model.Workout) {
	days := make(map[string]bool)
	for _, w := range workouts {
		days[w.StartDate.In(cstLoc).Format("2006-01-02")] = true
	}

	s.db.Transaction(func(tx *gorm.DB) error {
		for day := range days {
			dayStart, _ := time.ParseInLocation("2006-01-02", day, cstLoc)
			dayEnd := dayStart.Add(24 * time.Hour)
			tx.Where("start_date >= ? AND start_date < ?", dayStart, dayEnd).Delete(&model.Workout{})
		}
		return tx.CreateInBatches(workouts, 1000).Error
	})
}

func (s *ImportService) updateImportLog(importLog *model.ImportLog, recordCount int, status, errMsg string) {
	importLog.RecordCount = recordCount
	importLog.Status = status
	importLog.Error = errMsg
	s.db.Save(importLog)
}

package dal

import (
	"apple_health/biz/dal/postgres"
	"apple_health/bootstrao"
	"apple_health/utils/config"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	var gormLogger logger.Interface
	if config.Cfg.Server.LogLevel != "debug" {
		gormLogger = logger.Default.LogMode(logger.Error) // 只有错误日志
	} else {
		gormLogger = logger.Default.LogMode(logger.Info) // 输出信息级别的日志
	}

	DB = postgres.Init(config.Cfg.Db.User, config.Cfg.Db.Password, config.Cfg.Db.Host, config.Cfg.Db.Port, config.Cfg.Db.Database, config.Cfg.Server.Zone, gormLogger)
	err := bootstrao.Migrate(DB)
	if err != nil {
		return
	}

}

func ChackDb() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 调用 Ping 检查数据库连接是否正常
	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	return nil
}

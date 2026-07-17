package dal

import (
	"apple_health/biz/dal/postgres"
	"apple_health/bootstrao"
	"apple_health/utils/config"
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var Redis *redis.Client

func Init() {
	var gormLogger logger.Interface
	if config.Cfg.Server.LogLevel != "debug" {
		gormLogger = logger.Default.LogMode(logger.Error) // 只有错误日志
	} else {
		gormLogger = logger.Default.LogMode(logger.Info) // 输出信息级别的日志
	}

	DB = postgres.Init(config.Cfg.Db.User, config.Cfg.Db.Password, config.Cfg.Db.Host, config.Cfg.Db.Port, config.Cfg.Db.Database, config.Cfg.Server.Zone, gormLogger)
	initRedis()
	err := bootstrao.Migrate(DB)
	if err != nil {
		return
	}

}

func initRedis() {
	if !config.Cfg.Redis.Enabled() {
		return
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Cfg.Redis.Address(),
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("redis连接失败: %w", err))
	}

	Redis = client
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

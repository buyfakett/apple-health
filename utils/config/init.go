package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port          int    `mapstructure:"port"`
	Name          string `mapstructure:"name"`
	Author        string `mapstructure:"author"`
	LogLevel      string `mapstructure:"log_level"`
	EnableSwagger bool   `mapstructure:"swagger"`
	Zone          string `mapstructure:"zone"`
	Token         string `mapstructure:"token"`
}

type DbConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type AppConfig struct {
	Server ServerConfig `mapstructure:"server"`
	Db     DbConfig     `mapstructure:"db"`
	Admin  AdminConfig  `mapstructure:"admin"`
}

var Cfg AppConfig

func InitConfig(defaultConfigContent []byte) {
	// 1. 处理命令行参数
	ParseCLI()

	// 如果显示版本信息，直接退出
	if CliCfg.ShowVersion {
		return
	}

	v := viper.New()

	// 2. 加载嵌入的默认配置文件
	if len(defaultConfigContent) > 0 {
		v.SetConfigType("yaml")
		if err := v.ReadConfig(bytes.NewBuffer(defaultConfigContent)); err != nil {
			fmt.Printf("加载默认配置失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 3. 加载外部配置文件（如果存在）
	if CliCfg.ConfigFile != "" {
		if _, err := os.Stat(CliCfg.ConfigFile); err == nil {
			v.SetConfigFile(CliCfg.ConfigFile)
			if err := v.MergeInConfig(); err != nil {
				fmt.Printf("加载外部配置失败: %v (路径: %s)\n", err, CliCfg.ConfigFile)
				os.Exit(1)
			}
		} else {
			fmt.Printf("警告: 外部配置文件不存在，使用默认配置 (路径: %s)\n", CliCfg.ConfigFile)
		}
	}

	// 4. 环境变量覆盖
	v.SetEnvPrefix("apple_health")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 5. 合并命令行参数
	_ = v.BindPFlags(pflag.CommandLine)

	// 6. 映射到结构体
	if err := v.Unmarshal(&Cfg); err != nil {
		fmt.Println("解析配置失败:", err)
		os.Exit(1)
	}

	// 7. 设置默认值
	Cfg.Server.Name = ServerName
	Cfg.Server.Author = Author
}

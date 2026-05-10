package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// CLIConfig 表示命令行配置
type CLIConfig struct {
	ShowVersion bool
	ConfigFile  string
	Port        int
}

var CliCfg CLIConfig

// ParseCLI 解析命令行参数
func ParseCLI() {
	// 定义命令行参数
	pflag.BoolVarP(&CliCfg.ShowVersion, "version", "v", false, "显示版本信息")
	pflag.StringVarP(&CliCfg.ConfigFile, "config", "c", "", "配置文件路径")
	pflag.IntVarP(&CliCfg.Port, "port", "p", 8888, "服务端口")

	if (pflag.Lookup("help") != nil && pflag.Lookup("help").Value.String() == "true") || (len(os.Args) > 1 && os.Args[1] == "help") {
		pflag.PrintDefaults()
		os.Exit(0)
	}

	// 解析命令行参数
	pflag.Parse()
}

// ShowVersionAndExit 显示版本信息并退出
func ShowVersionAndExit(version string) {
	version = strings.TrimSpace(version)
	fmt.Printf("confkeeper %s\n", version)
	os.Exit(0)
}

package main

import (
	"apple_health/biz/dal"
	genrouter "apple_health/biz/router"
	"apple_health/utils/config"
	"apple_health/utils/logger"
	_ "embed"
	"fmt"

	"github.com/buyfakett/qingfeng"
	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
)

//go:embed config/default.yaml
var defaultConfigContent []byte

//go:embed docs/swagger.json
var swaggerJSoN []byte

//go:embed internal/version/version.txt
var version string

// @contact.name buyfakett
// @contact.url https://github.com/buyfakett

// @securityDefinitions.apikey	ApiKeyAuth
// @in	header
// @name authorization
func main() {
	config.InitConfig(defaultConfigContent)
	// 如果显示版本信息，直接退出
	if config.CliCfg.ShowVersion {
		config.ShowVersionAndExit(version)
	}
	logger.InitLog(config.Cfg.Server.LogLevel)
	if config.Cfg.Server.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	dal.Init()
	gin.ForceConsoleColor()
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	// 注册路由
	genrouter.RegisterRoutes(r)

	// 注册swagger文档
	if config.Cfg.Server.EnableSwagger {
		slog.Info("Swagger文档已启用")
		r.GET("/api/swagger/*any", qingfeng.Handler(qingfeng.Config{
			Version: version,
			Title:   config.Cfg.Server.Name,
			Description: fmt.Sprintf("%s by [%s](https://github.com/%s).",
				config.Cfg.Server.Name, config.Cfg.Server.Author, config.Cfg.Server.Author),
			DarkMode: true,
			BasePath: "/api/swagger",
			DocJSON:  swaggerJSoN,
			UITheme:  qingfeng.ThemeModern,
		}))
	}

	if config.Cfg.Server.LogLevel == "debug" {
		slog.Infof("服务启动成功，地址为 http://localhost:%d", config.Cfg.Server.Port)
	}

	r.NoRoute(func(c *gin.Context) { c.JSON(404, gin.H{"code": 404, "msg": "你访问的页面不存在"}) })

	// 启动服务
	port := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	if err := r.Run(port); err != nil {
		panic(err)
	}
}

package handler

import (
	"apple_health/biz/dal"
	"apple_health/internal/version"
	"apple_health/utils/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping 测试网络接口
// @Tags 测试
// @Summary 测试网络接口
// @Description 测试网络接口
// @Accept application/json
// @Produce application/json
// @Router /api/ping [get]
func Ping(c *gin.Context) {
	err := dal.ChackDb()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "数据库连接失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "pong",
	})
}

// ServerInfo 服务信息
// @Tags 测试
// @Summary 服务信息
// @Description 服务信息
// @Accept application/json
// @Produce application/json
// @Router /api/server_info [get]
func ServerInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": gin.H{
			"name":    config.Cfg.Server.Name,
			"version": version.Version,
		},
	})
}

package router

import (
	hHealth "apple_health/biz/handler/health"
	"apple_health/biz/mw"

	"github.com/gin-gonic/gin"
)

func healthRoutes(apiGroup *gin.RouterGroup) {
	healthGroup := apiGroup.Group("/health")
	{
		healthGroup.POST("/upload", mw.TokenAuthMiddleware(), hHealth.UploadHealthData)
		healthGroup.POST("/sync", mw.TokenAuthMiddleware(), hHealth.SyncHealthData)
		healthGroup.POST("/redis/rebuild", mw.TokenAuthMiddleware(), hHealth.RebuildHealthRedisCache)
		healthGroup.GET("/redis/daily", hHealth.GetHealthRedisCache)
	}
}

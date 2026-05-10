package router

import (
	hUser "apple_health/biz/handler/user"
	"apple_health/biz/mw"

	"github.com/gin-gonic/gin"
)

func userRoutes(apiGroup *gin.RouterGroup) {
	userGroup := apiGroup.Group("/user")
	{
		userGroup.GET("/test_token", mw.TokenAuthMiddleware(), hUser.TestToken)
	}
}

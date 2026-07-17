package mw

import (
	"apple_health/biz/response"
	"apple_health/utils/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenAuthMiddleware 鉴权中间件
func TokenAuthMiddleware() gin.HandlerFunc {
	return tokenAuthMiddleware()
}

func tokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization Header
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code": response.Code_Unauthorized,
				"msg":  "缺少token",
			})
			c.Abort() // 终止后续处理
			return
		}

		// 提取token（去除Bearer前缀）
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code": response.Code_Unauthorized,
				"msg":  "token格式错误",
			})
			c.Abort() // 终止后续处理
			return
		}

		if token != config.Cfg.Server.Token {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code": response.Code_Unauthorized,
				"msg":  "token错误",
			})
			c.Abort() // 终止后续处理
			return
		}

		// 如果验证通过，继续处理请求
		c.Next()
	}
}

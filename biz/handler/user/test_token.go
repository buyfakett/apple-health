package user

import (
	"apple_health/biz/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TestToken 测试token权限
//
//	@Tags			用户
//	@Summary		测试token权限
//	@Description	测试token权限
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	response.CommonResp
//	@Router			/api/user/test_token [get]
func TestToken(c *gin.Context) {
	c.JSON(http.StatusOK, &response.CommonResp{
		Code: response.Code_Success,
		Msg:  "测试成功",
	})
	return
}

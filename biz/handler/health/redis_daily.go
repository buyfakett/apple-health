package health

import (
	"apple_health/biz/service"
	"apple_health/utils/config"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GetHealthRedisCacheReq struct {
	Days int `form:"days" binding:"required,min=1,max=30" example:"7"`
}

type GetHealthRedisCacheResp struct {
	Code    int                       `json:"code" example:"200"`
	Message string                    `json:"message" example:"查询成功"`
	Data    *service.HealthCacheRange `json:"data,omitempty"`
}

// GetHealthRedisCache 读取Redis中的每日健康缓存
// @Summary 读取Redis中的每日健康缓存
// @Description 读取Redis中近days天的步数数组和详细健身记录数组；days最大30
// @Tags 健康数据
// @Produce json
// @Param days query int true "近N天，1-30"
// @Success 200 {object} GetHealthRedisCacheResp
// @Failure 400 {object} GetHealthRedisCacheResp
// @Failure 503 {object} GetHealthRedisCacheResp
// @Router /api/health/redis/daily [get]
func GetHealthRedisCache(c *gin.Context) {
	var req GetHealthRedisCacheReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, GetHealthRedisCacheResp{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误，请提供days，范围1-30",
		})
		return
	}

	end := todayInServerZone()
	start := end.AddDate(0, 0, -(req.Days - 1))

	result, err := service.GetDailyHealthCacheRange(c.Request.Context(), start, end)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrHealthCacheDisabled) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, GetHealthRedisCacheResp{
			Code:    status,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetHealthRedisCacheResp{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data:    result,
	})
}

func todayInServerZone() time.Time {
	loc, err := time.LoadLocation(config.Cfg.Server.Zone)
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

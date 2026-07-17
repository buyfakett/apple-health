package health

import (
	"apple_health/biz/service"
	"apple_health/utils/config"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RebuildHealthRedisCacheReq struct {
	StartDate string `json:"start_date" binding:"required,datetime=2006-01-02" example:"2024-01-01"`
	EndDate   string `json:"end_date" binding:"required,datetime=2006-01-02" example:"2024-01-31"`
}

type RebuildHealthRedisCacheResp struct {
	Code    int                              `json:"code" example:"200"`
	Message string                           `json:"message" example:"Redis缓存写入成功"`
	Data    *service.HealthCacheWriteSummary `json:"data,omitempty"`
}

// RebuildHealthRedisCache 从PG回填指定日期的Redis缓存
// @Summary 从PG回填指定日期的Redis缓存
// @Description 从PostgreSQL读取指定日期或日期范围的步数统计和锻炼记录，并写入Redis
// @Tags 健康数据
// @Accept application/json
// @Produce json
// @Param body body RebuildHealthRedisCacheReq true "日期范围"
// @Success 200 {object} RebuildHealthRedisCacheResp
// @Failure 400 {object} RebuildHealthRedisCacheResp
// @Failure 503 {object} RebuildHealthRedisCacheResp
// @Router /api/health/redis/rebuild [post]
func RebuildHealthRedisCache(c *gin.Context) {
	var req RebuildHealthRedisCacheReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RebuildHealthRedisCacheResp{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误，请提供start_date和end_date，格式为YYYY-MM-DD",
		})
		return
	}

	start, err := parseHealthCacheDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, RebuildHealthRedisCacheResp{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	end, err := parseHealthCacheDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, RebuildHealthRedisCacheResp{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}
	if start.After(end) {
		c.JSON(http.StatusBadRequest, RebuildHealthRedisCacheResp{
			Code:    http.StatusBadRequest,
			Message: "start_date不能晚于end_date",
		})
		return
	}

	summary, err := service.WriteDailyHealthCacheRange(c.Request.Context(), start, end)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrHealthCacheDisabled) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, RebuildHealthRedisCacheResp{
			Code:    status,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RebuildHealthRedisCacheResp{
		Code:    http.StatusOK,
		Message: "Redis缓存写入成功",
		Data:    summary,
	})
}

func parseHealthCacheDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("日期不能为空")
	}

	loc, err := time.LoadLocation(config.Cfg.Server.Zone)
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	if t, err := time.ParseInLocation("2006-01-02", value, loc); err == nil {
		return t, nil
	}

	return time.Time{}, errors.New("日期格式错误，请使用YYYY-MM-DD格式")
}

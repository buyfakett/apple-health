package health

import (
	"apple_health/biz/model"
	"apple_health/biz/service"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHealthDataReq struct {
	File *multipart.FileHeader `form:"file" binding:"required" json:"-"`
}

type ImportDataResp struct {
	Code    int              `json:"code" example:"200"`
	Message string           `json:"message" example:"导入成功"`
	Data    *model.ImportLog `json:"data,omitempty"`
}

// UploadHealthData 上传健康数据
// @Summary 上传健康数据
// @Description 上传Apple健康导出的zip文件
// @Tags 健康数据
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "zip文件"
// @Success 200 {object} ImportDataResp
// @Router /api/health/upload [post]
func UploadHealthData(c *gin.Context) {
	var req UploadHealthDataReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, ImportDataResp{
			Code:    http.StatusBadRequest,
			Message: "请上传文件",
		})
		return
	}

	if filepath.Ext(req.File.Filename) != ".zip" {
		c.JSON(http.StatusBadRequest, ImportDataResp{
			Code:    http.StatusBadRequest,
			Message: "只支持 .zip 文件",
		})
		return
	}

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, ImportDataResp{
			Code:    http.StatusInternalServerError,
			Message: "创建上传目录失败",
		})
		return
	}

	timestamp := time.Now().Format("20060102150405")
	filename := timestamp + "_" + req.File.Filename
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(req.File, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, ImportDataResp{
			Code:    http.StatusInternalServerError,
			Message: "保存文件失败",
		})
		return
	}

	importService := service.NewImportService()
	importLog, err := importService.ImportFromZip(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ImportDataResp{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ImportDataResp{
		Code:    http.StatusOK,
		Message: "导入成功",
		Data:    importLog,
	})
}

package salary

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// TaxUploadHandler 个税上传 HTTP 端点
type TaxUploadHandler struct {
	svc *Service
}

// NewTaxUploadHandler 创建个税上传 Handler
func NewTaxUploadHandler(svc *Service) *TaxUploadHandler {
	return &TaxUploadHandler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *TaxUploadHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	taxUpload := rg.Group("/salary/tax-upload", authMiddleware, middleware.RequireOrg)
	{
		taxUpload.POST("", h.UploadTax)
		taxUpload.POST("/confirm", h.ConfirmTax)
	}
}

// UploadTax 上传个税 Excel（预览匹配结果）
func (h *TaxUploadHandler) UploadTax(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		logger.SugarLogger.Debugw("UploadTax: year 参数错误", "year_str", yearStr)
		response.BadRequest(c, "year 参数错误")
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		logger.SugarLogger.Debugw("UploadTax: month 参数错误", "month_str", monthStr)
		response.BadRequest(c, "month 参数错误")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		logger.SugarLogger.Debugw("UploadTax: 文件缺失", "error", err.Error())
		response.BadRequest(c, "请上传文件")
		return
	}

	if file.Size > 5*1024*1024 {
		logger.SugarLogger.Debugw("UploadTax: 文件过大", "size", file.Size)
		response.BadRequest(c, "文件大小不能超过 5MB")
		return
	}

	src, err := file.Open()
	if err != nil {
		logger.SugarLogger.Debugw("UploadTax: 读取文件失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeEmployeeMatch, "读取文件失败")
		return
	}
	defer src.Close()

	fileBytes := make([]byte, file.Size)
	if _, err := src.Read(fileBytes); err != nil {
		logger.SugarLogger.Debugw("UploadTax: 读取文件内容失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeEmployeeMatch, "读取文件内容失败")
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("UploadTax: 调用", "org_id", orgID, "year", year, "month", month, "filename", file.Filename)

	result, err := h.svc.UploadTaxFile(orgID, year, month, fileBytes)
	if err != nil {
		logger.SugarLogger.Debugw("UploadTax: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, CodeEmployeeMatch, err.Error())
		return
	}

	response.Success(c, result)
}

// ConfirmTax 确认个税上传（批量更新）
func (h *TaxUploadHandler) ConfirmTax(c *gin.Context) {
	var req ConfirmTaxUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("ConfirmTax: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if len(req.MatchedRows) == 0 {
		response.BadRequest(c, "没有需要确认的数据")
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("ConfirmTax: 调用", "org_id", orgID, "user_id", userID, "year", req.Year, "month", req.Month, "count", len(req.MatchedRows))

	if err := h.svc.ConfirmTaxUpload(orgID, userID, req.Year, req.Month, req.MatchedRows); err != nil {
		logger.SugarLogger.Debugw("ConfirmTax: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeEmployeeMatch, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "工资表个税数据已更新，请重新核算",
		"count":   len(req.MatchedRows),
	})
}

// ConfirmTaxUploadRequest 确认个税上传请求
type ConfirmTaxUploadRequest struct {
	Year        int              `json:"year" binding:"required"`
	Month       int              `json:"month" binding:"required"`
	MatchedRows []TaxMatchedRow  `json:"matched_rows" binding:"required"`
}

package salary

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// PerformanceHandler 绩效系数 HTTP 端点
type PerformanceHandler struct {
	performanceSvc *PerformanceService
}

// NewPerformanceHandler 创建绩效系数 Handler
func NewPerformanceHandler(performanceSvc *PerformanceService) *PerformanceHandler {
	return &PerformanceHandler{performanceSvc: performanceSvc}
}

// RegisterRoutes 注册绩效系数路由
func (h *PerformanceHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	perf := rg.Group("/salary", authMiddleware)
	{
		perf.PUT("/performance", middleware.RequireRole("owner", "admin"), h.SetPerformance)
		perf.GET("/performance", h.GetPerformance)
	}
}

// SetPerformanceRequest 设置绩效系数请求
type SetPerformanceRequest struct {
	Year         int                  `json:"year" binding:"required,min=2000,max=2100"`
	Month        int                  `json:"month" binding:"required,min=1,max=12"`
	Coefficients []CoefficientInput   `json:"coefficients" binding:"required,min=1"`
}

// SetPerformance 批量设置绩效系数
func (h *PerformanceHandler) SetPerformance(c *gin.Context) {
	var req SetPerformanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("SetPerformance: 调用", "org_id", orgID, "user_id", userID, "count", len(req.Coefficients))

	if err := h.performanceSvc.SetCoefficient(orgID, userID, req.Year, req.Month, req.Coefficients); err != nil {
		logger.SugarLogger.Debugw("SetPerformance: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetPerformance 获取某月绩效系数
func (h *PerformanceHandler) GetPerformance(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		response.BadRequest(c, "year 参数错误")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "month 参数错误")
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("GetPerformance: 查询", "org_id", orgID, "year", year, "month", month)

	result, err := h.performanceSvc.GetCoefficients(orgID, year, month)
	if err != nil {
		logger.SugarLogger.Debugw("GetPerformance: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, result)
}

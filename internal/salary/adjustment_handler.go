package salary

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// AdjustmentHandler 调薪 HTTP 端点
type AdjustmentHandler struct {
	adjustmentSvc *AdjustmentService
}

// NewAdjustmentHandler 创建调薪 Handler
func NewAdjustmentHandler(adjustmentSvc *AdjustmentService) *AdjustmentHandler {
	return &AdjustmentHandler{adjustmentSvc: adjustmentSvc}
}

// RegisterRoutes 注册调薪路由
func (h *AdjustmentHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	adj := rg.Group("/salary", authMiddleware, middleware.RequireOrg)
	{
		adj.POST("/adjustment", middleware.RequireRole("owner", "admin"), h.CreateAdjustment)
		adj.POST("/mass-adjustment", middleware.RequireRole("owner", "admin"), h.CreateMassAdjustment)
		adj.POST("/adjustment/preview", middleware.RequireRole("owner", "admin"), h.Preview)
		adj.GET("/adjustments", h.GetAdjustmentList)
	}
}

// CreateAdjustment 单人调薪
func (h *AdjustmentHandler) CreateAdjustment(c *gin.Context) {
	var req AdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("CreateAdjustment: 调用", "org_id", orgID, "user_id", userID, "employee_id", req.EmployeeID)

	if err := h.adjustmentSvc.CreateAdjustment(orgID, userID, &req); err != nil {
		logger.SugarLogger.Debugw("CreateAdjustment: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, nil)
}

// CreateMassAdjustment 部门普调
func (h *AdjustmentHandler) CreateMassAdjustment(c *gin.Context) {
	var req MassAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("CreateMassAdjustment: 调用", "org_id", orgID, "user_id", userID, "departments", req.DepartmentIDs)

	if err := h.adjustmentSvc.CreateMassAdjustment(orgID, userID, &req); err != nil {
		logger.SugarLogger.Debugw("CreateMassAdjustment: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, nil)
}

// Preview 调薪预览
func (h *AdjustmentHandler) Preview(c *gin.Context) {
	var req AdjustmentPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("AdjustmentPreview: 调用", "org_id", orgID)

	result, err := h.adjustmentSvc.Preview(orgID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("AdjustmentPreview: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, result)
}

// GetAdjustmentList 调薪记录列表
func (h *AdjustmentHandler) GetAdjustmentList(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	effectiveMonth := c.Query("effective_month")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logger.SugarLogger.Debugw("GetAdjustmentList: 查询", "org_id", orgID, "month", effectiveMonth, "page", page)

	records, total, err := h.adjustmentSvc.GetAdjustmentList(orgID, effectiveMonth, page, pageSize)
	if err != nil {
		logger.SugarLogger.Debugw("GetAdjustmentList: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.PageSuccess(c, records, total, page, pageSize)
}

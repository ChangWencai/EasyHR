package position

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// PositionHandler 岗位 HTTP 端点
type PositionHandler struct {
	svc *Service
}

// NewPositionHandler 创建岗位 Handler
func NewPositionHandler(svc *Service) *PositionHandler {
	return &PositionHandler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *PositionHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)

	authGroup.POST("/positions", middleware.RequireRole("owner", "admin"), h.CreatePosition)
	authGroup.GET("/positions", h.ListPositions)
	authGroup.GET("/positions/select-options", h.GetSelectOptions)
	authGroup.PUT("/positions/:id", middleware.RequireRole("owner", "admin"), h.UpdatePosition)
	authGroup.DELETE("/positions/:id", middleware.RequireRole("owner", "admin"), h.DeletePosition)
}

// CreatePosition 创建岗位
func (h *PositionHandler) CreatePosition(c *gin.Context) {
	var req CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("CreatePosition: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	pos, err := h.svc.CreatePosition(orgID, userID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("CreatePosition: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, 20300, err.Error())
		return
	}

	response.Success(c, pos)
}

// ListPositions 岗位列表
func (h *PositionHandler) ListPositions(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	positions, err := h.svc.ListPositions(orgID)
	if err != nil {
		logger.SugarLogger.Debugw("ListPositions: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 20301, "查询岗位列表失败")
		return
	}

	response.Success(c, positions)
}

// GetSelectOptions 获取岗位下拉选项
func (h *PositionHandler) GetSelectOptions(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	// 解析 department_id 查询参数
	var deptID *int64
	deptIDStr := c.Query("department_id")
	if deptIDStr != "" {
		id, err := strconv.ParseInt(deptIDStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "无效的部门ID")
			return
		}
		deptID = &id
	}

	options, err := h.svc.GetSelectOptions(orgID, deptID)
	if err != nil {
		logger.SugarLogger.Debugw("GetSelectOptions: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 20302, "查询岗位选项失败")
		return
	}

	response.Success(c, options)
}

// UpdatePosition 更新岗位
func (h *PositionHandler) UpdatePosition(c *gin.Context) {
	var req UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdatePosition: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的岗位ID")
		return
	}

	pos, err := h.svc.UpdatePosition(orgID, userID, id, &req)
	if err != nil {
		logger.SugarLogger.Debugw("UpdatePosition: 失败", "error", err.Error(), "position_id", id)
		response.Error(c, http.StatusBadRequest, 20303, err.Error())
		return
	}

	response.Success(c, pos)
}

// DeletePosition 删除岗位
func (h *PositionHandler) DeletePosition(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的岗位ID")
		return
	}

	if err := h.svc.DeletePosition(orgID, id); err != nil {
		logger.SugarLogger.Debugw("DeletePosition: 失败", "error", err.Error(), "position_id", id)
		response.Error(c, http.StatusBadRequest, 20304, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "岗位已删除"})
}

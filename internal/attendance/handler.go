package attendance

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 考勤管理 HTTP 端点
type Handler struct {
	svc *AttendanceService
}

// NewHandler 创建 Handler
func NewHandler(svc *AttendanceService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	g := rg.Group("/attendance")
	g.Use(authMiddleware)

	// 打卡规则
	g.GET("/rule", h.GetRule)
	g.PUT("/rule", h.SaveRule)

	// 班次管理
	g.GET("/shifts", h.ListShifts)
	g.POST("/shifts", h.CreateShift)
	g.PUT("/shifts/:id", h.UpdateShift)
	g.DELETE("/shifts/:id", h.DeleteShift)

	// 排班管理
	g.GET("/schedules", h.ListSchedules)
	g.POST("/schedules", h.BatchUpsertSchedules)
}

func getOrgID(c *gin.Context) int64  { return c.GetInt64("org_id") }
func getUserID(c *gin.Context) int64 { return c.GetInt64("user_id") }

// GetRule 获取打卡规则
func (h *Handler) GetRule(c *gin.Context) {
	rule, err := h.svc.GetRule(c.Request.Context(), getOrgID(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// SaveRule 保存打卡规则
func (h *Handler) SaveRule(c *gin.Context) {
	var req SaveAttendanceRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	rule, err := h.svc.SaveRule(c.Request.Context(), getOrgID(c), getUserID(c), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// ListShifts 获取班次列表
func (h *Handler) ListShifts(c *gin.Context) {
	shifts, err := h.svc.ListShifts(c.Request.Context(), getOrgID(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shifts})
}

// CreateShift 创建班次
func (h *Handler) CreateShift(c *gin.Context) {
	var req CreateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	shift, err := h.svc.CreateShift(c.Request.Context(), getOrgID(c), getUserID(c), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": shift})
}

// UpdateShift 更新班次
func (h *Handler) UpdateShift(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req UpdateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	shift, err := h.svc.UpdateShift(c.Request.Context(), getOrgID(c), getUserID(c), id, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shift})
}

// DeleteShift 删除班次
func (h *Handler) DeleteShift(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.DeleteShift(c.Request.Context(), getOrgID(c), id); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ListSchedules 获取排班列表
func (h *Handler) ListSchedules(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	employeeIDStr := c.Query("employee_id")
	var employeeID *int64
	if employeeIDStr != "" {
		if id, err := strconv.ParseInt(employeeIDStr, 10, 64); err == nil {
			employeeID = &id
		}
	}
	schedules, err := h.svc.ListSchedules(c.Request.Context(), getOrgID(c), startDate, endDate, employeeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

// BatchUpsertSchedules 批量保存排班
func (h *Handler) BatchUpsertSchedules(c *gin.Context) {
	var req BatchScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.BatchUpsertSchedules(c.Request.Context(), getOrgID(c), getUserID(c), &req); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "排班保存成功"})
}

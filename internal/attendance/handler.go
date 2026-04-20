package attendance

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 考勤管理 HTTP 端点
type Handler struct {
	svc         *AttendanceService
	approvalSvc *ApprovalService
}

// NewHandler 创建 Handler
func NewHandler(svc *AttendanceService) *Handler {
	return &Handler{svc: svc}
}

// SetApprovalService 注入审批服务（避免循环依赖）
func (h *Handler) SetApprovalService(approvalSvc *ApprovalService) {
	h.approvalSvc = approvalSvc
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

	// 打卡实况
	g.GET("/clock-live", h.GetClockLive)
	g.POST("/clock-records", h.CreateClockRecord)
	g.GET("/leave-stats", h.GetLeaveStats)
	g.PUT("/leave-stats/:employee_id", h.UpdateLeaveStats)

	// 审批流
	g.GET("/approvals/pending-count", h.GetPendingCount)
	g.GET("/approvals", h.ListApprovals)
	g.POST("/approvals", h.CreateApproval)
	g.PUT("/approvals/:id/approve", h.ApproveApproval)
	g.PUT("/approvals/:id/reject", h.RejectApproval)
	g.PUT("/approvals/:id/cancel", h.CancelApproval)

	// 出勤月报
	g.GET("/monthly", h.GetMonthlyReport)
	g.GET("/monthly/export", h.ExportMonthlyExcel)
	g.GET("/daily-records", h.GetDailyRecords)

	// 合规报表
	g.GET("/compliance/overtime", h.GetComplianceOvertime)
	g.GET("/compliance/leave", h.GetComplianceLeave)
	g.GET("/compliance/anomaly", h.GetComplianceAnomaly)
	g.GET("/compliance/monthly", h.GetComplianceMonthly)
	g.GET("/compliance/monthly/export", h.ExportComplianceMonthly)
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

// GetClockLive 获取今日打卡实况
func (h *Handler) GetClockLive(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	var departmentID *int64
	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		if id, err := strconv.ParseInt(deptIDStr, 10, 64); err == nil {
			departmentID = &id
		}
	}

	result, err := h.svc.GetClockLive(c.Request.Context(), getOrgID(c), date, departmentID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// CreateClockRecord 创建打卡记录（管理员代打/邀请点签）
func (h *Handler) CreateClockRecord(c *gin.Context) {
	var req CreateClockRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	record, err := h.svc.CreateClockRecord(c.Request.Context(), getOrgID(c), getUserID(c), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": record})
}

// GetLeaveStats 获取假勤统计
func (h *Handler) GetLeaveStats(c *gin.Context) {
	employeeIDStr := c.Query("employee_id")
	yearMonth := c.DefaultQuery("year_month", time.Now().Format("2006-01"))
	if employeeIDStr == "" {
		response.BadRequest(c, "employee_id required")
		return
	}
	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "employee_id 格式错误")
		return
	}
	stats, err := h.svc.GetLeaveStats(c.Request.Context(), getOrgID(c), employeeID, yearMonth)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// UpdateLeaveStats 手动修正假勤统计数据
func (h *Handler) UpdateLeaveStats(c *gin.Context) {
	employeeID, err := strconv.ParseInt(c.Param("employee_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "employee_id 格式错误")
		return
	}
	yearMonth := c.DefaultQuery("year_month", time.Now().Format("2006-01"))
	var req UpdateLeaveStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.UpdateLeaveStats(c.Request.Context(), getOrgID(c), getUserID(c), employeeID, yearMonth, &req); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// === Approval Handlers ===

// GetPendingCount 获取待审批条数
func (h *Handler) GetPendingCount(c *gin.Context) {
	if h.approvalSvc == nil {
		response.Error(c, http.StatusInternalServerError, 500, "审批服务未初始化")
		return
	}
	count, err := h.approvalSvc.GetPendingCount(c.Request.Context(), getOrgID(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": PendingCountResponse{PendingCount: count}})
}

// ListApprovals 审批列表
func (h *Handler) ListApprovals(c *gin.Context) {
	if h.approvalSvc == nil {
		response.Error(c, http.StatusInternalServerError, 500, "审批服务未初始化")
		return
	}
	status := c.Query("status")
	approvalType := c.Query("approval_type")
	employeeIDStr := c.Query("employee_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var employeeID *int64
	if employeeIDStr != "" {
		if id, err := strconv.ParseInt(employeeIDStr, 10, 64); err == nil {
			employeeID = &id
		}
	}

	result, err := h.approvalSvc.ListApprovals(c.Request.Context(), getOrgID(c), status, approvalType, employeeID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// CreateApproval 创建审批申请
func (h *Handler) CreateApproval(c *gin.Context) {
	if h.approvalSvc == nil {
		response.Error(c, http.StatusInternalServerError, 500, "审批服务未初始化")
		return
	}
	var req CreateApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.approvalSvc.CreateApproval(c.Request.Context(), getOrgID(c), getUserID(c), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// ApproveApproval 审批通过
func (h *Handler) ApproveApproval(c *gin.Context) {
	if h.approvalSvc == nil {
		response.Error(c, http.StatusInternalServerError, 500, "审批服务未初始化")
		return
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	result, err := h.approvalSvc.Approve(c.Request.Context(), getOrgID(c), getUserID(c), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// RejectApproval 审批驳回
func (h *Handler) RejectApproval(c *gin.Context) {
	if h.approvalSvc == nil {
		response.Error(c, http.StatusInternalServerError, 500, "审批服务未初始化")
		return
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req RejectApprovalRequest
	_ = c.ShouldBindJSON(&req) // note 为选填
	result, err := h.approvalSvc.Reject(c.Request.Context(), getOrgID(c), getUserID(c), id, req.Note)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// CancelApproval 撤回申请
func (h *Handler) CancelApproval(c *gin.Context) {
	if h.approvalSvc == nil {
		response.Error(c, http.StatusInternalServerError, 500, "审批服务未初始化")
		return
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	result, err := h.approvalSvc.Cancel(c.Request.Context(), getOrgID(c), getUserID(c), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetMonthlyReport 出勤月报
func (h *Handler) GetMonthlyReport(c *gin.Context) {
	yearMonth := c.DefaultQuery("year_month", time.Now().Format("2006-01"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	result, err := h.svc.GetMonthlyReport(c.Request.Context(), getOrgID(c), yearMonth, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ExportMonthlyExcel 导出月报 Excel
func (h *Handler) ExportMonthlyExcel(c *gin.Context) {
	yearMonth := c.DefaultQuery("year_month", time.Now().Format("2006-01"))
	data, filename, err := h.svc.ExportMonthlyExcel(c.Request.Context(), getOrgID(c), yearMonth)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// GetDailyRecords 每日打卡详情
func (h *Handler) GetDailyRecords(c *gin.Context) {
	employeeIDStr := c.Query("employee_id")
	yearMonth := c.DefaultQuery("year_month", time.Now().Format("2006-01"))
	if employeeIDStr == "" {
		response.BadRequest(c, "employee_id required")
		return
	}
	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "employee_id 格式错误")
		return
	}
	result, err := h.svc.GetDailyRecords(c.Request.Context(), getOrgID(c), employeeID, yearMonth)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// === Compliance Report Handlers ===

func (h *Handler) GetComplianceOvertime(c *gin.Context) {
	var req ComplianceReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "year_month 参数必填，格式: YYYY-MM")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	resp, err := h.svc.GetComplianceOvertime(c.Request.Context(), getOrgID(c), &req, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) GetComplianceLeave(c *gin.Context) {
	var req ComplianceReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "year_month 参数必填，格式: YYYY-MM")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	resp, err := h.svc.GetComplianceLeave(c.Request.Context(), getOrgID(c), &req, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) GetComplianceAnomaly(c *gin.Context) {
	var req ComplianceReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "year_month 参数必填，格式: YYYY-MM")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	resp, err := h.svc.GetComplianceAnomaly(c.Request.Context(), getOrgID(c), &req, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) GetComplianceMonthly(c *gin.Context) {
	var req ComplianceReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "year_month 参数必填，格式: YYYY-MM")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	resp, err := h.svc.GetComplianceMonthly(c.Request.Context(), getOrgID(c), &req, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *Handler) ExportComplianceMonthly(c *gin.Context) {
	var req ComplianceReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 400, "year_month 参数必填，格式: YYYY-MM")
		return
	}
	data, filename, err := h.svc.ExportComplianceMonthlyExcel(c.Request.Context(), getOrgID(c), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename*=UTF-8''%s`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

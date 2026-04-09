package salary

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 工资核算 HTTP 端点
type Handler struct {
	svc *Service
}

// NewHandler 创建工资核算 Handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	salary := rg.Group("/salary", authMiddleware)
	{
		// 薪资模板管理
		salary.GET("/template", h.GetTemplate)
		salary.PUT("/template", middleware.RequireRole("owner", "admin"), h.UpdateTemplate)

		// 员工薪资项管理
		salary.GET("/items", h.GetEmployeeItems)
		salary.PUT("/items/:employee_id", middleware.RequireRole("owner", "admin"), h.SetEmployeeItems)

		// 工资核算流程
		salary.POST("/payroll", middleware.RequireRole("owner", "admin"), h.CreatePayroll)
		salary.GET("/payroll", h.GetPayrollList)
		salary.GET("/payroll/:id", h.GetPayrollDetail)
		salary.POST("/payroll/calculate", middleware.RequireRole("owner", "admin"), h.CalculatePayroll)
		salary.PUT("/payroll/confirm", middleware.RequireRole("owner", "admin"), h.ConfirmPayroll)
		salary.PUT("/payroll/:id/pay", middleware.RequireRole("owner", "admin"), h.RecordPayment)

		// 考勤导入
		salary.POST("/attendance/import", middleware.RequireRole("owner", "admin"), h.ImportAttendance)
	}
}

// GetTemplate 获取企业薪资模板
func (h *Handler) GetTemplate(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	tmpl, err := h.svc.GetTemplate(orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, tmpl)
}

// UpdateTemplate 更新企业薪资模板
func (h *Handler) UpdateTemplate(c *gin.Context) {
	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.UpdateTemplate(orgID, userID, req.Items); err != nil {
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetEmployeeItems 获取员工薪资项
func (h *Handler) GetEmployeeItems(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	employeeIDStr := c.Query("employee_id")
	month := c.Query("month")

	if employeeIDStr == "" || month == "" {
		response.BadRequest(c, "employee_id 和 month 为必填参数")
		return
	}

	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "employee_id 格式错误")
		return
	}

	items, err := h.svc.GetEmployeeItems(orgID, employeeID, month)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, items)
}

// SetEmployeeItems 设置员工薪资项
func (h *Handler) SetEmployeeItems(c *gin.Context) {
	employeeIDStr := c.Param("employee_id")
	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "employee_id 格式错误")
		return
	}

	var req SetEmployeeItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.SetEmployeeItems(orgID, userID, employeeID, req.Month, req.Items); err != nil {
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, fmt.Sprintf("员工 %d 薪资项已更新", employeeID))
}

// ========== 工资核算流程 Handler ==========

// CreatePayroll 创建月度工资表
func (h *Handler) CreatePayroll(c *gin.Context) {
	var req CreatePayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	records, err := h.svc.CreatePayroll(orgID, userID, req.Year, req.Month, req.CopyFromMonth)
	if err != nil {
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, records)
}

// GetPayrollList 查询工资表列表
func (h *Handler) GetPayrollList(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	yearStr := c.Query("year")
	monthStr := c.Query("month")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

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

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := h.svc.GetPayrollList(orgID, year, month, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.PageSuccess(c, records, total, page, pageSize)
}

// GetPayrollDetail 查询单个工资记录详情
func (h *Handler) GetPayrollDetail(c *gin.Context) {
	idStr := c.Param("id")
	recordID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "id 格式错误")
		return
	}

	orgID := c.GetInt64("org_id")

	record, err := h.svc.GetPayrollDetail(orgID, recordID)
	if err != nil {
		response.Error(c, http.StatusNotFound, CodePayrollNotFound, err.Error())
		return
	}

	response.Success(c, record)
}

// CalculatePayroll 一键核算
func (h *Handler) CalculatePayroll(c *gin.Context) {
	var req struct {
		Year  int `json:"year" binding:"required,min=2000,max=2100"`
		Month int `json:"month" binding:"required,min=1,max=12"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.CalculatePayroll(orgID, userID, req.Year, req.Month)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, result)
}

// ConfirmPayroll 确认工资表
func (h *Handler) ConfirmPayroll(c *gin.Context) {
	var req ConfirmPayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.ConfirmPayroll(orgID, userID, req.Year, req.Month)
	if err != nil {
		response.Error(c, http.StatusBadRequest, CodeInvalidStatus, err.Error())
		return
	}

	response.Success(c, result)
}

// RecordPayment 发放记录
func (h *Handler) RecordPayment(c *gin.Context) {
	idStr := c.Param("id")
	recordID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "id 格式错误")
		return
	}

	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.RecordPayment(orgID, userID, recordID, &req); err != nil {
		response.Error(c, http.StatusBadRequest, CodeInvalidStatus, err.Error())
		return
	}

	response.Success(c, nil)
}

// ImportAttendance 考勤 Excel 导入
func (h *Handler) ImportAttendance(c *gin.Context) {
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

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}

	if file.Size > 5*1024*1024 { // 5MB 限制
		response.BadRequest(c, "文件大小不能超过 5MB")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "读取文件失败")
		return
	}
	defer src.Close()

	fileBytes := make([]byte, file.Size)
	if _, err := src.Read(fileBytes); err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "读取文件内容失败")
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.ImportAttendance(orgID, userID, year, month, fileBytes)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, result)
}

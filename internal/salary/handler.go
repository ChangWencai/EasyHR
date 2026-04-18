package salary

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 工资核算 HTTP 端点
type Handler struct {
	svc          *Service
	dashboardSvc *SalaryDashboardService
}

// NewHandler 创建工资核算 Handler
func NewHandler(svc *Service, dashboardSvc *SalaryDashboardService) *Handler {
	return &Handler{svc: svc, dashboardSvc: dashboardSvc}
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

		// 薪资看板
		salary.GET("/dashboard", h.GetDashboard)

		// 考勤导入
		salary.POST("/attendance/import", middleware.RequireRole("owner", "admin"), h.ImportAttendance)

		// 工资单推送
		salary.POST("/slip/send", middleware.RequireRole("owner", "admin"), h.SendSlip)

		// 导出
		salary.GET("/payroll/export", h.ExportPayroll)
	}

	// 公开端点（H5 工资单查看，无需认证）
	public := rg.Group("/salary/slip")
	{
		public.GET("/:token", h.GetSlipByToken)
		public.POST("/:token/verify", h.VerifySlipPhone)
		public.POST("/:token/code", h.VerifySlipCode)
		public.POST("/:token/sign", h.SignSlip)
	}
}

// GetTemplate 获取企业薪资模板
func (h *Handler) GetTemplate(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("GetTemplate: 查询", "org_id", orgID)

	tmpl, err := h.svc.GetTemplate(orgID)
	if err != nil {
		logger.SugarLogger.Debugw("GetTemplate: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, tmpl)
}

// UpdateTemplate 更新企业薪资模板
func (h *Handler) UpdateTemplate(c *gin.Context) {
	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdateTemplate: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("UpdateTemplate: 调用", "org_id", orgID, "user_id", userID)

	if err := h.svc.UpdateTemplate(orgID, userID, req.Items); err != nil {
		logger.SugarLogger.Debugw("UpdateTemplate: 失败", "error", err.Error(), "org_id", orgID)
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
		logger.SugarLogger.Debugw("GetEmployeeItems: 参数缺失", "employee_id", employeeIDStr, "month", month)
		response.BadRequest(c, "employee_id 和 month 为必填参数")
		return
	}

	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		logger.SugarLogger.Debugw("GetEmployeeItems: employee_id 格式错误", "error", err.Error(), "employee_id_str", employeeIDStr)
		response.BadRequest(c, "employee_id 格式错误")
		return
	}

	logger.SugarLogger.Debugw("GetEmployeeItems: 查询", "org_id", orgID, "employee_id", employeeID, "month", month)
	items, err := h.svc.GetEmployeeItems(orgID, employeeID, month)
	if err != nil {
		logger.SugarLogger.Debugw("GetEmployeeItems: 失败", "error", err.Error(), "org_id", orgID, "employee_id", employeeID)
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
		logger.SugarLogger.Debugw("SetEmployeeItems: employee_id 格式错误", "error", err.Error(), "employee_id_str", employeeIDStr)
		response.BadRequest(c, "employee_id 格式错误")
		return
	}

	var req SetEmployeeItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("SetEmployeeItems: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("SetEmployeeItems: 调用", "org_id", orgID, "user_id", userID, "employee_id", employeeID, "month", req.Month)

	if err := h.svc.SetEmployeeItems(orgID, userID, employeeID, req.Month, req.Items); err != nil {
		logger.SugarLogger.Debugw("SetEmployeeItems: 失败", "error", err.Error(), "org_id", orgID, "employee_id", employeeID)
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
		logger.SugarLogger.Debugw("CreatePayroll: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("CreatePayroll: 调用", "org_id", orgID, "user_id", userID, "year", req.Year, "month", req.Month, "copy_from", req.CopyFromMonth)

	records, err := h.svc.CreatePayroll(orgID, userID, req.Year, req.Month, req.CopyFromMonth)
	if err != nil {
		logger.SugarLogger.Debugw("CreatePayroll: 失败", "error", err.Error(), "org_id", orgID, "year", req.Year, "month", req.Month)
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
		logger.SugarLogger.Debugw("GetPayrollList: year 参数错误", "year_str", yearStr)
		response.BadRequest(c, "year 参数错误")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		logger.SugarLogger.Debugw("GetPayrollList: month 参数错误", "month_str", monthStr)
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

	logger.SugarLogger.Debugw("GetPayrollList: 查询", "org_id", orgID, "year", year, "month", month, "page", page, "page_size", pageSize)
	records, total, err := h.svc.GetPayrollList(orgID, year, month, page, pageSize)
	if err != nil {
		logger.SugarLogger.Debugw("GetPayrollList: 失败", "error", err.Error(), "org_id", orgID)
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
		logger.SugarLogger.Debugw("GetPayrollDetail: id 格式错误", "error", err.Error(), "id_str", idStr)
		response.BadRequest(c, "id 格式错误")
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("GetPayrollDetail: 查询", "org_id", orgID, "record_id", recordID)

	record, err := h.svc.GetPayrollDetail(orgID, recordID)
	if err != nil {
		logger.SugarLogger.Debugw("GetPayrollDetail: 失败", "error", err.Error(), "org_id", orgID, "record_id", recordID)
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
		logger.SugarLogger.Debugw("CalculatePayroll: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("CalculatePayroll: 调用", "org_id", orgID, "user_id", userID, "year", req.Year, "month", req.Month)

	result, err := h.svc.CalculatePayroll(orgID, userID, req.Year, req.Month)
	if err != nil {
		logger.SugarLogger.Debugw("CalculatePayroll: 失败", "error", err.Error(), "org_id", orgID, "year", req.Year, "month", req.Month)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, result)
}

// ConfirmPayroll 确认工资表
func (h *Handler) ConfirmPayroll(c *gin.Context) {
	var req ConfirmPayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("ConfirmPayroll: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("ConfirmPayroll: 调用", "org_id", orgID, "user_id", userID, "year", req.Year, "month", req.Month)

	result, err := h.svc.ConfirmPayroll(orgID, userID, req.Year, req.Month)
	if err != nil {
		logger.SugarLogger.Debugw("ConfirmPayroll: 失败", "error", err.Error(), "org_id", orgID, "year", req.Year, "month", req.Month)
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
		logger.SugarLogger.Debugw("RecordPayment: id 格式错误", "error", err.Error(), "id_str", idStr)
		response.BadRequest(c, "id 格式错误")
		return
	}

	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("RecordPayment: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("RecordPayment: 调用", "org_id", orgID, "user_id", userID, "record_id", recordID)

	if err := h.svc.RecordPayment(orgID, userID, recordID, &req); err != nil {
		logger.SugarLogger.Debugw("RecordPayment: 失败", "error", err.Error(), "org_id", orgID, "record_id", recordID)
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
		logger.SugarLogger.Debugw("ImportAttendance: year 参数错误", "year_str", yearStr)
		response.BadRequest(c, "year 参数错误")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		logger.SugarLogger.Debugw("ImportAttendance: month 参数错误", "month_str", monthStr)
		response.BadRequest(c, "month 参数错误")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		logger.SugarLogger.Debugw("ImportAttendance: 文件缺失", "error", err.Error())
		response.BadRequest(c, "请上传文件")
		return
	}

	if file.Size > 5*1024*1024 { // 5MB 限制
		logger.SugarLogger.Debugw("ImportAttendance: 文件过大", "size", file.Size)
		response.BadRequest(c, "文件大小不能超过 5MB")
		return
	}

	src, err := file.Open()
	if err != nil {
		logger.SugarLogger.Debugw("ImportAttendance: 读取文件失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "读取文件失败")
		return
	}
	defer src.Close()

	fileBytes := make([]byte, file.Size)
	if _, err := src.Read(fileBytes); err != nil {
		logger.SugarLogger.Debugw("ImportAttendance: 读取文件内容失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "读取文件内容失败")
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("ImportAttendance: 调用", "org_id", orgID, "user_id", userID, "year", year, "month", month, "filename", file.Filename)

	result, err := h.svc.ImportAttendance(orgID, userID, year, month, fileBytes)
	if err != nil {
		logger.SugarLogger.Debugw("ImportAttendance: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, result)
}

// ========== 工资单 Handler ==========

// SendSlip 发送工资单
func (h *Handler) SendSlip(c *gin.Context) {
	var req SendSlipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("SendSlip: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("SendSlip: 调用", "org_id", orgID, "user_id", userID, "record_ids", req.RecordIDs)

	results, err := h.svc.SendSlip(orgID, userID, req.RecordIDs)
	if err != nil {
		logger.SugarLogger.Debugw("SendSlip: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, results)
}

// GetSlipByToken 获取工资单详情（公开端点，无需认证）
func (h *Handler) GetSlipByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		logger.SugarLogger.Debugw("GetSlipByToken: token 参数错误")
		response.BadRequest(c, "token 参数错误")
		return
	}

	logger.SugarLogger.Debugw("GetSlipByToken: 查询", "token", token)
	detail, err := h.svc.GetSlipByToken(token)
	if err != nil {
		if err == ErrSlipTokenInvalid {
			logger.SugarLogger.Debugw("GetSlipByToken: token 无效", "token", token)
			response.Error(c, http.StatusNotFound, CodePayrollNotFound, "工资单不存在")
			return
		}
		if err == ErrSlipTokenExpired {
			logger.SugarLogger.Debugw("GetSlipByToken: token 已过期", "token", token)
			response.Error(c, http.StatusForbidden, CodePayrollNotFound, "工资单已过期")
			return
		}
		logger.SugarLogger.Debugw("GetSlipByToken: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, detail)
}

// VerifySlipPhone 验证工资单手机号（发送短信验证码）
func (h *Handler) VerifySlipPhone(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		logger.SugarLogger.Debugw("VerifySlipPhone: token 参数错误")
		response.BadRequest(c, "token 参数错误")
		return
	}

	var req VerifySlipPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("VerifySlipPhone: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	logger.SugarLogger.Debugw("VerifySlipPhone: 调用", "token", token, "phone", req.Phone)
	if err := h.svc.VerifySlipPhone(token, req.Phone); err != nil {
		if err == ErrSlipTokenInvalid {
			logger.SugarLogger.Debugw("VerifySlipPhone: token 无效", "token", token)
			response.Error(c, http.StatusNotFound, CodePayrollNotFound, "工资单不存在")
			return
		}
		if err == ErrSlipTokenExpired {
			logger.SugarLogger.Debugw("VerifySlipPhone: token 已过期", "token", token)
			response.Error(c, http.StatusForbidden, CodePayrollNotFound, "工资单已过期")
			return
		}
		logger.SugarLogger.Debugw("VerifySlipPhone: 失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, nil)
}

// VerifySlipCode 验证短信验证码
func (h *Handler) VerifySlipCode(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		logger.SugarLogger.Debugw("VerifySlipCode: token 参数错误")
		response.BadRequest(c, "token 参数错误")
		return
	}

	var req VerifySlipCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("VerifySlipCode: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	logger.SugarLogger.Debugw("VerifySlipCode: 调用", "token", token, "phone", req.Phone)
	valid, err := h.svc.VerifySlipCode(token, req.Phone, req.Code)
	if err != nil {
		if err == ErrSlipTokenInvalid {
			logger.SugarLogger.Debugw("VerifySlipCode: token 无效", "token", token)
			response.Error(c, http.StatusNotFound, CodePayrollNotFound, "工资单不存在")
			return
		}
		if err == ErrSlipTokenExpired {
			logger.SugarLogger.Debugw("VerifySlipCode: token 已过期", "token", token)
			response.Error(c, http.StatusForbidden, CodePayrollNotFound, "工资单已过期")
			return
		}
		logger.SugarLogger.Debugw("VerifySlipCode: 失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, gin.H{"valid": valid})
}

// SignSlip 签收工资单
func (h *Handler) SignSlip(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		logger.SugarLogger.Debugw("SignSlip: token 参数错误")
		response.BadRequest(c, "token 参数错误")
		return
	}

	logger.SugarLogger.Debugw("SignSlip: 调用", "token", token)
	if err := h.svc.SignSlip(token); err != nil {
		if err == ErrSlipTokenInvalid {
			logger.SugarLogger.Debugw("SignSlip: token 无效", "token", token)
			response.Error(c, http.StatusNotFound, CodePayrollNotFound, "工资单不存在")
			return
		}
		if err == ErrSlipTokenExpired {
			logger.SugarLogger.Debugw("SignSlip: token 已过期", "token", token)
			response.Error(c, http.StatusForbidden, CodePayrollNotFound, "工资单已过期")
			return
		}
		if err == ErrSlipAlreadySigned {
			logger.SugarLogger.Debugw("SignSlip: 已签收", "token", token)
			response.Error(c, http.StatusBadRequest, CodeInvalidStatus, "工资单已签收")
			return
		}
		logger.SugarLogger.Debugw("SignSlip: 失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, nil)
}

// ExportPayroll 导出工资表 Excel
func (h *Handler) ExportPayroll(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		logger.SugarLogger.Debugw("ExportPayroll: year 参数错误", "year_str", yearStr)
		response.BadRequest(c, "year 参数错误")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		logger.SugarLogger.Debugw("ExportPayroll: month 参数错误", "month_str", monthStr)
		response.BadRequest(c, "month 参数错误")
		return
	}

	// 查询工资记录
	logger.SugarLogger.Debugw("ExportPayroll: 查询", "org_id", orgID, "year", year, "month", month)
	records, err := h.svc.repo.FindPayrollRecordsByMonth(orgID, year, month)
	if err != nil {
		logger.SugarLogger.Debugw("ExportPayroll: 查询工资记录失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "查询工资记录失败")
		return
	}

	// 查询工资明细
	var recordsWithItems []PayrollRecordWithItems
	for _, record := range records {
		items, err := h.svc.repo.FindPayrollItemsByRecord(orgID, record.ID)
		if err != nil {
			continue
		}
		recordsWithItems = append(recordsWithItems, PayrollRecordWithItems{
			Record: record,
			Items:  items,
		})
	}

	// 生成 Excel
	data, err := ExportPayrollExcel(recordsWithItems, year, month)
	if err != nil {
		logger.SugarLogger.Debugw("ExportPayroll: 生成 Excel 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "生成 Excel 失败")
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=工资条_%d_%02d.xlsx", year, month))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

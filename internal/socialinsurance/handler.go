package socialinsurance

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 社保 HTTP 端点
type Handler struct {
	svc          *Service
	dashboardSvc *SIDashboardService
	paymentRepo  *SIMonthlyPaymentRepository
}

// NewHandler 创建社保 Handler
func NewHandler(svc *Service, dashboardSvc *SIDashboardService, paymentRepo *SIMonthlyPaymentRepository) *Handler {
	return &Handler{svc: svc, dashboardSvc: dashboardSvc, paymentRepo: paymentRepo}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)

	// 政策管理仅 OWNER 可操作（D-14），查询和计算所有角色可访问
	authGroup.POST("/social-insurance/policies", middleware.RequireRole("owner"), h.CreatePolicy)
	authGroup.GET("/social-insurance/policies", h.ListPolicies)
	authGroup.GET("/social-insurance/policies/:id", h.GetPolicy)
	authGroup.PUT("/social-insurance/policies/:id", middleware.RequireRole("owner"), h.UpdatePolicy)
	authGroup.DELETE("/social-insurance/policies/:id", middleware.RequireRole("owner"), h.DeletePolicy)
	authGroup.POST("/social-insurance/calculate", h.CalculateInsurance)

	// 参保操作（OWNER/ADMIN）
	authGroup.POST("/social-insurance/enroll/preview", middleware.RequireRole("owner", "admin"), h.EnrollPreview)
	authGroup.POST("/social-insurance/enroll", middleware.RequireRole("owner", "admin"), h.BatchEnroll)
	authGroup.POST("/social-insurance/stop", middleware.RequireRole("owner", "admin"), h.BatchStopEnrollment)

	// 查询（所有角色，MEMBER 通过 GetMyRecords 只看自己）
	authGroup.GET("/social-insurance/records", h.ListRecords)
	authGroup.GET("/social-insurance/my-records", h.GetMyRecords)
	authGroup.GET("/social-insurance/records/:id/history", h.GetChangeHistory)
	authGroup.GET("/social-insurance/deduction", h.GetDeduction)

	// Phase 8 新增路由（D-SI-05~D-SI-16）
	authGroup.GET("/social-insurance/dashboard", h.SIDashboard)
	authGroup.POST("/social-insurance/enroll/single", middleware.RequireRole("owner", "admin"), h.Enroll)
	authGroup.POST("/social-insurance/stop/single", middleware.RequireRole("owner", "admin"), h.Stop)
	authGroup.POST("/social-insurance/payment-callback", h.PaymentCallback)
	authGroup.GET("/social-insurance/monthly-records", h.GetMonthlyRecords)
	authGroup.GET("/social-insurance/monthly-records/:id", h.GetMonthlyRecordDetail)
	authGroup.POST("/social-insurance/confirm-payment", middleware.RequireRole("owner", "admin"), h.ConfirmPayment)
}

// CreatePolicy 创建社保政策
func (h *Handler) CreatePolicy(c *gin.Context) {
	var req CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("CreatePolicy: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("CreatePolicy: 创建", "user_id", userID, "city_id", req.CityID)

	policy := &SocialInsurancePolicy{
		CityID:        req.CityID,
		EffectiveYear: req.EffectiveYear,
		Config:        newJSONType(req.Config),
	}
	policy.CreatedBy = userID
	policy.UpdatedBy = userID

	if err := h.svc.CreatePolicy(policy); err != nil {
		logger.SugarLogger.Debugw("CreatePolicy: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, 30100, "创建社保政策失败")
		return
	}

	response.Success(c, h.svc.toPolicyResponse(policy))
}

// GetPolicy 获取政策详情
func (h *Handler) GetPolicy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.SugarLogger.Debugw("GetPolicy: 无效ID", "param", c.Param("id"))
		response.BadRequest(c, "无效的政策ID")
		return
	}

	logger.SugarLogger.Debugw("GetPolicy: 查询", "policy_id", id)
	policy, err := h.svc.GetPolicy(id)
	if err != nil {
		logger.SugarLogger.Debugw("GetPolicy: 失败", "error", err.Error(), "policy_id", id)
		response.Error(c, http.StatusNotFound, 30101, err.Error())
		return
	}

	response.Success(c, policy)
}

// ListPolicies 政策列表
func (h *Handler) ListPolicies(c *gin.Context) {
	var query PolicyListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.SugarLogger.Debugw("ListPolicies: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 默认分页参数
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	logger.SugarLogger.Debugw("ListPolicies: 查询", "city_id", query.CityID, "page", query.Page)
	policies, total, err := h.svc.ListPolicies(query.CityID, query.Page, query.PageSize)
	if err != nil {
		logger.SugarLogger.Debugw("ListPolicies: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, 30102, "查询政策列表失败")
		return
	}

	response.PageSuccess(c, policies, total, query.Page, query.PageSize)
}

// UpdatePolicy 更新政策
func (h *Handler) UpdatePolicy(c *gin.Context) {
	var req UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdatePolicy: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.SugarLogger.Debugw("UpdatePolicy: 无效ID", "param", c.Param("id"))
		response.BadRequest(c, "无效的政策ID")
		return
	}

	logger.SugarLogger.Debugw("UpdatePolicy: 更新", "policy_id", id)
	if err := h.svc.UpdatePolicy(id, &req); err != nil {
		logger.SugarLogger.Debugw("UpdatePolicy: 失败", "error", err.Error(), "policy_id", id)
		response.Error(c, http.StatusBadRequest, 30103, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "政策已更新"})
}

// DeletePolicy 删除政策
func (h *Handler) DeletePolicy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.SugarLogger.Debugw("DeletePolicy: 无效ID", "param", c.Param("id"))
		response.BadRequest(c, "无效的政策ID")
		return
	}

	logger.SugarLogger.Debugw("DeletePolicy: 删除", "policy_id", id)
	if err := h.svc.DeletePolicy(id); err != nil {
		logger.SugarLogger.Debugw("DeletePolicy: 失败", "error", err.Error(), "policy_id", id)
		response.Error(c, http.StatusBadRequest, 30104, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "政策已删除"})
}

// CalculateInsurance 社保金额计算
func (h *Handler) CalculateInsurance(c *gin.Context) {
	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("CalculateInsurance: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	logger.SugarLogger.Debugw("CalculateInsurance: 计算", "city_id", req.CityID, "salary", req.Salary)
	result, err := h.svc.CalculateInsuranceAmounts(req.CityID, req.Salary, req.Year)
	if err != nil {
		logger.SugarLogger.Debugw("CalculateInsurance: 失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, 30105, err.Error())
		return
	}

	response.Success(c, result)
}

// ========== 参保/停缴/查询端点 ==========

// EnrollPreview 参保预览
func (h *Handler) EnrollPreview(c *gin.Context) {
	var req EnrollPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("EnrollPreview: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("EnrollPreview: 预览", "org_id", orgID)

	result, err := h.svc.EnrollPreview(orgID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("EnrollPreview: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 30200, "参保预览失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// BatchEnroll 批量参保
func (h *Handler) BatchEnroll(c *gin.Context) {
	var req BatchEnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("BatchEnroll: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("BatchEnroll: 批量参保", "org_id", orgID, "user_id", userID)

	result, err := h.svc.BatchEnroll(orgID, userID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("BatchEnroll: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 30201, "批量参保失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// BatchStopEnrollment 批量停缴
func (h *Handler) BatchStopEnrollment(c *gin.Context) {
	var req BatchStopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("BatchStopEnrollment: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("BatchStopEnrollment: 批量停缴", "org_id", orgID, "user_id", userID)

	result, err := h.svc.BatchStopEnrollment(orgID, userID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("BatchStopEnrollment: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 30202, "批量停缴失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// ListRecords 参保记录列表
func (h *Handler) ListRecords(c *gin.Context) {
	var params RecordListQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.SugarLogger.Debugw("ListRecords: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("ListRecords: 查询", "org_id", orgID)

	records, total, page, pageSize, err := h.svc.ListRecords(orgID, params)
	if err != nil {
		logger.SugarLogger.Debugw("ListRecords: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 30203, "查询参保记录失败")
		return
	}

	response.PageSuccess(c, records, total, page, pageSize)
}

// GetMyRecords 查询自己的社保记录（MEMBER 角色）
func (h *Handler) GetMyRecords(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	logger.SugarLogger.Debugw("GetMyRecords: 查询", "org_id", orgID, "user_id", userID)
	records, err := h.svc.GetMyRecords(orgID, userID)
	if err != nil {
		logger.SugarLogger.Debugw("GetMyRecords: 失败", "error", err.Error())
		response.Error(c, http.StatusNotFound, 30204, err.Error())
		return
	}

	response.Success(c, records)
}

// GetChangeHistory 变更历史查询
func (h *Handler) GetChangeHistory(c *gin.Context) {
	var query struct {
		EmployeeID int64 `form:"employee_id"`
		Page       int   `form:"page"`
		PageSize   int   `form:"page_size"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.SugarLogger.Debugw("GetChangeHistory: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("GetChangeHistory: 查询", "org_id", orgID, "employee_id", query.EmployeeID)

	histories, total, err := h.svc.GetChangeHistory(orgID, query.EmployeeID, query.Page, query.PageSize)
	if err != nil {
		logger.SugarLogger.Debugw("GetChangeHistory: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, 30205, "查询变更历史失败")
		return
	}

	response.PageSuccess(c, histories, total, query.Page, query.PageSize)
}

// GetDeduction 社保扣款查询（Phase 5 调用）
func (h *Handler) GetDeduction(c *gin.Context) {
	var query struct {
		EmployeeID int64  `form:"employee_id" binding:"required"`
		Month      string `form:"month" binding:"required"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.SugarLogger.Debugw("GetDeduction: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("GetDeduction: 查询", "org_id", orgID, "employee_id", query.EmployeeID, "month", query.Month)

	deduction, err := h.svc.GetSocialInsuranceDeduction(orgID, query.EmployeeID, query.Month)
	if err != nil {
		logger.SugarLogger.Debugw("GetDeduction: 失败", "error", err.Error())
		response.Error(c, http.StatusNotFound, 30206, err.Error())
		return
	}

	response.Success(c, deduction)
}

// ========== Phase 8 新增 Handler ==========

// SIDashboard 社保数据看板（SI-01~SI-04）
func (h *Handler) SIDashboard(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	yearMonth := c.DefaultQuery("year_month", "")

	if yearMonth == "" {
		// 默认当月
		now := time.Now()
		yearMonth = fmt.Sprintf("%04d%02d", now.Year(), now.Month())
	}

	logger.SugarLogger.Debugw("SIDashboard: 查询", "org_id", orgID, "year_month", yearMonth)

	result, err := h.dashboardSvc.GetDashboard(c.Request.Context(), orgID, yearMonth)
	if err != nil {
		logger.SugarLogger.Debugw("SIDashboard: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 30300, "查询社保看板失败")
		return
	}

	response.Success(c, result)
}

// Enroll 单个增员（SI-05~SI-08）
func (h *Handler) Enroll(c *gin.Context) {
	var req EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("Enroll: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("Enroll: 增员", "org_id", orgID, "employee_id", req.EmployeeID)

	// 转换为 BatchEnrollRequest 复用现有逻辑
	batchReq := &BatchEnrollRequest{
		EmployeeIDs: []int64{req.EmployeeID},
		CityID:      req.CityID,
		StartMonth:  req.StartYearMonth,
	}

	result, err := h.svc.BatchEnroll(orgID, userID, batchReq)
	if err != nil {
		logger.SugarLogger.Debugw("Enroll: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 30301, "增员失败: "+err.Error())
		return
	}

	// 为该员工创建当月 SIMonthlyPayment 记录（status=pending）
	if result.SuccessCount > 0 {
		yearMonth := req.StartYearMonth
		if len(yearMonth) == 7 {
			// YYYY-MM -> YYYYMM
			yearMonth = yearMonth[:4] + yearMonth[5:]
		}
		if yearMonth == "" {
			now := time.Now()
			yearMonth = fmt.Sprintf("%04d%02d", now.Year(), now.Month())
		}

		// 从参保记录获取缴费金额
		record, rErr := h.svc.repo.FindActiveRecordByEmployee(orgID, req.EmployeeID)
		if rErr == nil && record != nil {
			payment := &SIMonthlyPayment{
				EmployeeID:     uint(req.EmployeeID),
				YearMonth:      yearMonth,
				Status:         PaymentStatusPending,
				PaymentChannel: SIPayChannelSelf,
				CompanyAmount:  decimal.NewFromFloat(record.TotalCompany),
				PersonalAmount: decimal.NewFromFloat(record.TotalPersonal),
				TotalAmount:    decimal.NewFromFloat(record.TotalCompany + record.TotalPersonal),
			}
			payment.OrgID = orgID
			payment.CreatedBy = userID
			payment.UpdatedBy = userID

			_ = h.paymentRepo.Create(c.Request.Context(), nil, payment)
		}
	}

	response.Success(c, result)
}

// Stop 单个减员（SI-09~SI-13）
func (h *Handler) Stop(c *gin.Context) {
	var req StopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("Stop: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("Stop: 减员", "org_id", orgID, "employee_id", req.EmployeeID)

	// 查找该员工的 active 参保记录
	record, err := h.svc.repo.FindActiveRecordByEmployee(orgID, req.EmployeeID)
	if err != nil {
		logger.SugarLogger.Debugw("Stop: 未找到参保记录", "error", err.Error(), "employee_id", req.EmployeeID)
		response.Error(c, http.StatusNotFound, 30302, "未找到该员工的参保记录")
		return
	}

	// 转换为 BatchStopRequest 复用现有逻辑
	stopYearMonth := req.StopYearMonth
	if len(stopYearMonth) == 7 {
		// YYYY-MM -> YYYY-MM（保持原格式，与 SocialInsuranceRecord.end_month 一致）
	}

	batchReq := &BatchStopRequest{
		RecordIDs: []int64{record.ID},
		EndMonth:  stopYearMonth,
	}

	result, err := h.svc.BatchStopEnrollment(orgID, userID, batchReq)
	if err != nil {
		logger.SugarLogger.Debugw("Stop: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 30303, "减员失败: "+err.Error())
		return
	}

	// 更新该员工的 SIMonthlyPayment 状态为 transferred（D-SI-03）
	if result.SuccessCount > 0 {
		ym := stopYearMonth
		if len(ym) == 7 {
			ym = ym[:4] + ym[5:]
		}
		payment, pErr := h.paymentRepo.GetByOrgAndEmployee(c.Request.Context(), orgID, uint(req.EmployeeID), ym)
		if pErr == nil && payment != nil {
			_ = h.paymentRepo.UpdateStatus(c.Request.Context(), orgID, payment.ID, PaymentStatusTransferred)
		}
	}

	response.Success(c, result)
}

// PaymentCallback 代理缴费 webhook（SI-16）
func (h *Handler) PaymentCallback(c *gin.Context) {
	var req PaymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("PaymentCallback: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// TODO: HMAC 签名验证（T-08-04），当前先验证 payment_id 存在
	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("PaymentCallback: 回调", "payment_id", req.PaymentID, "status", req.Status)

	var newStatus PaymentStatus
	if req.Status == "success" {
		newStatus = PaymentStatusNormal
	} else {
		newStatus = PaymentStatusPending // 失败可重试
	}

	if err := h.paymentRepo.UpdateStatus(c.Request.Context(), orgID, req.PaymentID, newStatus); err != nil {
		logger.SugarLogger.Debugw("PaymentCallback: 失败", "error", err.Error(), "payment_id", req.PaymentID)
		response.Error(c, http.StatusInternalServerError, 30304, "更新缴费状态失败")
		return
	}

	response.Success(c, gin.H{"message": "状态已更新"})
}

// GetMonthlyRecords 参保记录列表（SI-11/SI-18，含月度缴费状态）
func (h *Handler) GetMonthlyRecords(c *gin.Context) {
	var params RecordListQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		logger.SugarLogger.Debugw("GetMonthlyRecords: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 20
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("GetMonthlyRecords: 查询", "org_id", orgID)

	// 复用现有参保记录查询
	records, total, page, pageSize, err := h.svc.ListRecords(orgID, params)
	if err != nil {
		logger.SugarLogger.Debugw("GetMonthlyRecords: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 30305, "查询参保记录失败")
		return
	}

	response.PageSuccess(c, records, total, page, pageSize)
}

// GetMonthlyRecordDetail 五险分项详情（SI-19/SI-20）
func (h *Handler) GetMonthlyRecordDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.SugarLogger.Debugw("GetMonthlyRecordDetail: 无效ID", "param", c.Param("id"))
		response.BadRequest(c, "无效的记录ID")
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("GetMonthlyRecordDetail: 查询", "org_id", orgID, "id", id)

	// 查询参保记录获取五险分项
	record, err := h.svc.repo.FindRecordByID(orgID, id)
	if err != nil {
		logger.SugarLogger.Debugw("GetMonthlyRecordDetail: 失败", "error", err.Error(), "id", id)
		response.Error(c, http.StatusNotFound, 30306, "参保记录不存在")
		return
	}

	response.Success(c, record)
}

// ConfirmPayment 自主缴费确认（SI-15）
func (h *Handler) ConfirmPayment(c *gin.Context) {
	var req ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("ConfirmPayment: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("ConfirmPayment: 确认", "org_id", orgID, "payment_id", req.PaymentID)

	if err := h.paymentRepo.UpdateStatus(c.Request.Context(), orgID, req.PaymentID, PaymentStatusNormal); err != nil {
		logger.SugarLogger.Debugw("ConfirmPayment: 失败", "error", err.Error(), "payment_id", req.PaymentID)
		response.Error(c, http.StatusInternalServerError, 30307, "确认缴费失败")
		return
	}

	response.Success(c, gin.H{"message": "缴费已确认"})
}

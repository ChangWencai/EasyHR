package socialinsurance

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// Handler 社保 HTTP 端点
type Handler struct {
	svc *Service
}

// NewHandler 创建社保 Handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
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
}

// CreatePolicy 创建社保政策
func (h *Handler) CreatePolicy(c *gin.Context) {
	var req CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt64("user_id")

	policy := &SocialInsurancePolicy{
		CityID:        req.CityID,
		EffectiveYear: req.EffectiveYear,
		Config:        newJSONType(req.Config),
	}
	policy.CreatedBy = userID
	policy.UpdatedBy = userID

	if err := h.svc.CreatePolicy(policy); err != nil {
		response.Error(c, http.StatusInternalServerError, 30100, "创建社保政策失败")
		return
	}

	response.Success(c, h.svc.toPolicyResponse(policy))
}

// GetPolicy 获取政策详情
func (h *Handler) GetPolicy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的政策ID")
		return
	}

	policy, err := h.svc.GetPolicy(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, 30101, err.Error())
		return
	}

	response.Success(c, policy)
}

// ListPolicies 政策列表
func (h *Handler) ListPolicies(c *gin.Context) {
	var query PolicyListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
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

	policies, total, err := h.svc.ListPolicies(query.CityID, query.Page, query.PageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 30102, "查询政策列表失败")
		return
	}

	response.PageSuccess(c, policies, total, query.Page, query.PageSize)
}

// UpdatePolicy 更新政策
func (h *Handler) UpdatePolicy(c *gin.Context) {
	var req UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的政策ID")
		return
	}

	if err := h.svc.UpdatePolicy(id, &req); err != nil {
		response.Error(c, http.StatusBadRequest, 30103, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "政策已更新"})
}

// DeletePolicy 删除政策
func (h *Handler) DeletePolicy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的政策ID")
		return
	}

	if err := h.svc.DeletePolicy(id); err != nil {
		response.Error(c, http.StatusBadRequest, 30104, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "政策已删除"})
}

// CalculateInsurance 社保金额计算
func (h *Handler) CalculateInsurance(c *gin.Context) {
	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.svc.CalculateInsuranceAmounts(req.CityID, req.Salary, req.Year)
	if err != nil {
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
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	result, err := h.svc.EnrollPreview(orgID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 30200, "参保预览失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// BatchEnroll 批量参保
func (h *Handler) BatchEnroll(c *gin.Context) {
	var req BatchEnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.BatchEnroll(orgID, userID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 30201, "批量参保失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// BatchStopEnrollment 批量停缴
func (h *Handler) BatchStopEnrollment(c *gin.Context) {
	var req BatchStopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	result, err := h.svc.BatchStopEnrollment(orgID, userID, &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 30202, "批量停缴失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// ListRecords 参保记录列表
func (h *Handler) ListRecords(c *gin.Context) {
	var params RecordListQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	records, total, page, pageSize, err := h.svc.ListRecords(orgID, params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 30203, "查询参保记录失败")
		return
	}

	response.PageSuccess(c, records, total, page, pageSize)
}

// GetMyRecords 查询自己的社保记录（MEMBER 角色）
func (h *Handler) GetMyRecords(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	records, err := h.svc.GetMyRecords(orgID, userID)
	if err != nil {
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

	histories, total, err := h.svc.GetChangeHistory(orgID, query.EmployeeID, query.Page, query.PageSize)
	if err != nil {
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
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	deduction, err := h.svc.GetSocialInsuranceDeduction(orgID, query.EmployeeID, query.Month)
	if err != nil {
		response.Error(c, http.StatusNotFound, 30206, err.Error())
		return
	}

	response.Success(c, deduction)
}

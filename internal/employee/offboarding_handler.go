package employee

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
	"gorm.io/datatypes"
)

// OffboardingHandler 离职管理 HTTP 端点
type OffboardingHandler struct {
	svc *OffboardingService
}

// NewOffboardingHandler 创建离职 Handler
func NewOffboardingHandler(svc *OffboardingService) *OffboardingHandler {
	return &OffboardingHandler{svc: svc}
}

// RegisterRoutes 注册离职管理路由
func (h *OffboardingHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)

	// 老板办理离职 — OWNER/ADMIN
	authGroup.POST("/employees/:id/resign", middleware.RequireRole("owner", "admin"), h.BossResign)
	// 员工申请离职（H5，需认证）
	authGroup.POST("/employees/:id/resign/apply", h.EmployeeApplyResign)
	// 审批离职申请 — OWNER/ADMIN
	authGroup.PUT("/offboardings/:id/approve", middleware.RequireRole("owner", "admin"), h.ApproveResign)
	// 驳回离职申请 — OWNER/ADMIN
	authGroup.PUT("/offboardings/:id/reject", middleware.RequireRole("owner", "admin"), h.RejectResign)
	// 完成交接 — OWNER/ADMIN
	authGroup.PUT("/offboardings/:id/complete", middleware.RequireRole("owner", "admin"), h.CompleteOffboarding)
	// 离职详情 — 所有角色
	authGroup.GET("/offboardings/:id", h.GetOffboarding)
	// 更新交接清单 — OWNER/ADMIN
	authGroup.PUT("/offboardings/:id/checklist", middleware.RequireRole("owner", "admin"), h.UpdateChecklist)
	// 离职列表 — 所有角色
	authGroup.GET("/offboardings", h.ListOffboardings)
}

// BossResign 老板直接办理离职
func (h *OffboardingHandler) BossResign(c *gin.Context) {
	employeeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	var req BossResignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.BossResign(orgID, userID, employeeID, &req); err != nil {
		response.Error(c, http.StatusBadRequest, 20200, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "离职办理成功"})
}

// EmployeeApplyResign 员工申请离职
func (h *OffboardingHandler) EmployeeApplyResign(c *gin.Context) {
	employeeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	var req EmployeeResignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	if err := h.svc.EmployeeApplyResign(orgID, employeeID, &req); err != nil {
		response.Error(c, http.StatusBadRequest, 20201, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "离职申请已提交"})
}

// ApproveResign 审批离职申请
func (h *OffboardingHandler) ApproveResign(c *gin.Context) {
	offboardingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的离职记录ID")
		return
	}

	orgID := c.GetInt64("org_id")
	approverID := c.GetInt64("user_id")

	if err := h.svc.ApproveResign(orgID, approverID, offboardingID); err != nil {
		response.Error(c, http.StatusBadRequest, 20202, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "审批通过"})
}

// RejectResign 驳回离职申请
func (h *OffboardingHandler) RejectResign(c *gin.Context) {
	offboardingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的离职记录ID")
		return
	}

	var req RejectResignRequest
	// reason 是选填的，不需要严格绑定错误处理
	_ = c.ShouldBindJSON(&req)

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	if err := h.svc.RejectResign(orgID, userID, offboardingID, req.Reason); err != nil {
		response.Error(c, http.StatusBadRequest, 20210, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "已驳回"})
}

// CompleteOffboarding 完成交接
func (h *OffboardingHandler) CompleteOffboarding(c *gin.Context) {
	offboardingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的离职记录ID")
		return
	}

	orgID := c.GetInt64("org_id")

	if err := h.svc.CompleteOffboarding(orgID, offboardingID); err != nil {
		response.Error(c, http.StatusBadRequest, 20203, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "交接完成"})
}

// GetOffboarding 获取离职详情
func (h *OffboardingHandler) GetOffboarding(c *gin.Context) {
	offboardingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的离职记录ID")
		return
	}

	orgID := c.GetInt64("org_id")

	detail, err := h.svc.GetOffboarding(orgID, offboardingID)
	if err != nil {
		response.Error(c, http.StatusNotFound, 20204, err.Error())
		return
	}

	response.Success(c, detail)
}

// UpdateChecklist 更新交接清单
func (h *OffboardingHandler) UpdateChecklist(c *gin.Context) {
	offboardingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的离职记录ID")
		return
	}

	var req UpdateChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	if err := h.svc.UpdateChecklist(orgID, offboardingID, req.ChecklistItems); err != nil {
		response.Error(c, http.StatusBadRequest, 20205, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "交接清单已更新"})
}

// ListOffboardings 离职列表
func (h *OffboardingHandler) ListOffboardings(c *gin.Context) {
	var query OffboardingListQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")

	offboardings, total, err := h.svc.ListOffboardings(orgID, query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 20206, "查询离职列表失败")
		return
	}

	response.PageSuccess(c, offboardings, total, query.Page, query.PageSize)
}

// ensure datatypes import is used
var _ datatypes.JSON

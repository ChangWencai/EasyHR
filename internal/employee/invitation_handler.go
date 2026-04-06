package employee

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// InvitationHandler 邀请 HTTP 端点
type InvitationHandler struct {
	svc *InvitationService
}

// NewInvitationHandler 创建邀请 Handler
func NewInvitationHandler(svc *InvitationService) *InvitationHandler {
	return &InvitationHandler{svc: svc}
}

// RegisterRoutes 注册路由（公开接口 + 认证接口混合）
func (h *InvitationHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)

	// 公开接口（无需认证）— 员工填写信息
	rg.GET("/invitations/:token", h.GetInvitationDetail)
	rg.POST("/invitations/:token/submit", h.SubmitInvitation)

	// 需要认证的接口 — 老板操作
	authGroup.POST("/invitations", middleware.RequireRole("owner", "admin"), h.CreateInvitation)
	authGroup.GET("/invitations", h.ListInvitations)
	authGroup.DELETE("/invitations/:id", middleware.RequireRole("owner", "admin"), h.CancelInvitation)
	authGroup.POST("/employees/:id/confirm", middleware.RequireRole("owner", "admin"), h.ConfirmOnboarding)
}

// CreateInvitation 创建邀请
func (h *InvitationHandler) CreateInvitation(c *gin.Context) {
	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	resp, err := h.svc.CreateInvitation(orgID, userID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20200, err.Error())
		return
	}

	response.Success(c, resp)
}

// ListInvitations 邀请列表
func (h *InvitationHandler) ListInvitations(c *gin.Context) {
	var query ListInvitationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	items, total, err := h.svc.ListInvitations(orgID, query)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 20201, "查询邀请列表失败")
		return
	}

	response.PageSuccess(c, items, total, query.Page, query.PageSize)
}

// GetInvitationDetail 获取邀请详情（公开接口）
func (h *InvitationHandler) GetInvitationDetail(c *gin.Context) {
	token := c.Param("token")

	detail, err := h.svc.GetInvitationDetail(token)
	if err != nil {
		if err == ErrInvitationNotFound {
			response.Error(c, http.StatusNotFound, 20202, err.Error())
			return
		}
		if err == ErrInvitationExpired {
			response.Error(c, http.StatusGone, 20203, err.Error())
			return
		}
		if err == ErrInvitationUsed || err == ErrInvitationCancelled {
			response.Error(c, http.StatusBadRequest, 20204, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, 20205, "查询邀请详情失败")
		return
	}

	response.Success(c, detail)
}

// SubmitInvitation 员工提交信息（公开接口）
func (h *InvitationHandler) SubmitInvitation(c *gin.Context) {
	token := c.Param("token")

	var req SubmitInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.SubmitInvitation(token, &req); err != nil {
		if err == ErrInvitationNotFound {
			response.Error(c, http.StatusNotFound, 20206, err.Error())
			return
		}
		if err == ErrInvitationExpired {
			response.Error(c, http.StatusGone, 20207, err.Error())
			return
		}
		if err == ErrInvitationUsed || err == ErrInvitationCancelled {
			response.Error(c, http.StatusBadRequest, 20208, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 20209, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "提交成功"})
}

// CancelInvitation 取消邀请
func (h *InvitationHandler) CancelInvitation(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的邀请ID")
		return
	}

	if err := h.svc.CancelInvitation(orgID, id); err != nil {
		if err == ErrInvitationNotFound {
			response.Error(c, http.StatusNotFound, 20210, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 20211, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "邀请已取消"})
}

// ConfirmOnboarding 确认入职
func (h *InvitationHandler) ConfirmOnboarding(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的员工ID")
		return
	}

	if err := h.svc.ConfirmOnboarding(orgID, id); err != nil {
		if err == ErrEmployeeNotPending {
			response.Error(c, http.StatusBadRequest, 20212, err.Error())
			return
		}
		response.Error(c, http.StatusNotFound, 20213, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "入职确认成功"})
}

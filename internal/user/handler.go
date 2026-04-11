package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)

	rg.POST("/auth/send-code", h.SendCode)
	rg.POST("/auth/login", h.Login)
	rg.POST("/auth/login/password", h.LoginPassword) // 新增：密码登录
	rg.POST("/auth/refresh", h.Refresh)
	authGroup.POST("/auth/logout", h.Logout)
	authGroup.GET("/auth/me", h.GetMe) // 新增：获取当前用户信息
	authGroup.PUT("/org/onboarding", h.CompleteOnboarding)

	authGroup.GET("/users", middleware.RequireRole("owner", "admin"), h.ListSubAccounts)
	authGroup.POST("/users", middleware.RequireRole("owner"), h.CreateSubAccount)
	authGroup.PUT("/users/:id/role", middleware.RequireRole("owner"), h.UpdateSubAccountRole)
	authGroup.DELETE("/users/:id", middleware.RequireRole("owner"), h.DeleteSubAccount)
}

func (h *Handler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.SendCode(c.Request.Context(), req.Phone); err != nil {
		response.Error(c, http.StatusTooManyRequests, 10002, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "验证码已发送"})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		if strings.Contains(err.Error(), "MEMBER_ROLE_FORBIDDEN") {
			response.Error(c, http.StatusForbidden, 10010, "您的账号为员工账号，请使用员工端微信小程序登录")
			return
		}
		response.Error(c, http.StatusUnauthorized, 10003, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	resp, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "刷新失败: "+err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")[7:]
	var req RefreshRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.Logout(c.Request.Context(), accessToken, req.RefreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, 10004, "退出失败")
		return
	}
	response.Success(c, gin.H{"message": "已退出登录"})
}

func (h *Handler) CompleteOnboarding(c *gin.Context) {
	var req CompleteOnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	orgID := c.GetInt64("org_id")
	if err := h.svc.CompleteOnboarding(c.Request.Context(), orgID, &req); err != nil {
		response.Error(c, http.StatusInternalServerError, 10005, "企业信息录入失败")
		return
	}
	response.Success(c, gin.H{"message": "企业信息录入成功"})
}

func (h *Handler) ListSubAccounts(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	users, total, err := h.svc.ListSubAccounts(c.Request.Context(), orgID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 10006, "查询用户列表失败")
		return
	}
	response.PageSuccess(c, users, total, page, pageSize)
}

func (h *Handler) CreateSubAccount(c *gin.Context) {
	var req CreateSubAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	orgID := c.GetInt64("org_id")
	if err := h.svc.CreateSubAccount(c.Request.Context(), orgID, &req); err != nil {
		response.Error(c, http.StatusBadRequest, 10007, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "子账号创建成功"})
}

func (h *Handler) UpdateSubAccountRole(c *gin.Context) {
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	orgID := c.GetInt64("org_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	if err := h.svc.UpdateSubAccountRole(c.Request.Context(), orgID, targetID, req.Role); err != nil {
		response.Error(c, http.StatusBadRequest, 10008, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "角色更新成功"})
}

func (h *Handler) DeleteSubAccount(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}
	if err := h.svc.DeleteSubAccount(c.Request.Context(), orgID, targetID); err != nil {
		response.Error(c, http.StatusBadRequest, 10009, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "用户已删除"})
}

func (h *Handler) LoginPassword(c *gin.Context) {
	var req PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.svc.LoginPassword(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "MEMBER_ROLE_FORBIDDEN") {
			// MEMBER 角色返回 403（per D-06, D-19, D-20）
			response.Error(c, http.StatusForbidden, 10010, "您的账号为员工账号，请使用员工端微信小程序登录")
			return
		}
		if strings.Contains(err.Error(), "该账号未设置密码") {
			response.Error(c, http.StatusUnauthorized, 10011, err.Error())
			return
		}
		response.Error(c, http.StatusUnauthorized, 10012, "手机号或密码错误")
		return
	}
	response.Success(c, resp)
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := c.GetInt64("user_id")
	orgID := c.GetInt64("org_id")

	resp, err := h.svc.GetMe(c.Request.Context(), userID, orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 10013, "获取用户信息失败")
		return
	}
	response.Success(c, resp)
}

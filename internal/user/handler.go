package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
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
	rg.POST("/auth/send-code", h.SendCode)
	rg.POST("/auth/login", h.Login)
	rg.POST("/auth/register", h.Register)
	rg.POST("/auth/login/password", h.LoginPassword)
	rg.POST("/auth/refresh", h.Refresh)

	// 认证但不需要 org 的路由
	authOnlyGroup := rg.Group("")
	authOnlyGroup.Use(authMiddleware)
	authOnlyGroup.PUT("/auth/org/onboarding", h.CompleteOnboarding)

	// 认证 + 需要 org 的路由
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware, middleware.RequireOrg)
	authGroup.POST("/auth/logout", h.Logout)
	authGroup.GET("/auth/me", h.GetMe)
	authGroup.PUT("/auth/password", h.ChangePassword)
	authGroup.PUT("/auth/avatar", h.UpdateAvatar)
	authGroup.PUT("/auth/name", h.UpdateName)
	authGroup.PUT("/org", h.UpdateOrg)
	authGroup.GET("/users", middleware.RequireRole("owner", "admin"), h.ListSubAccounts)
	authGroup.POST("/users", middleware.RequireRole("owner"), h.CreateSubAccount)
	authGroup.PUT("/users/:id/role", middleware.RequireRole("owner"), h.UpdateSubAccountRole)
	authGroup.DELETE("/users/:id", middleware.RequireRole("owner"), h.DeleteSubAccount)
}

func (h *Handler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("SendCode: 参数错误", "error", err.Error(), "phone", req.Phone)
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	logger.SugarLogger.Debugw("SendCode: 请求参数", "phone", req.Phone)
	if err := h.svc.SendCode(c.Request.Context(), req.Phone); err != nil {
		logger.SugarLogger.Debugw("SendCode: 发送失败", "error", err.Error(), "phone", req.Phone)
		response.Error(c, http.StatusTooManyRequests, 10002, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "验证码已发送"})
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("Register: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	logger.SugarLogger.Debugw("Register: 请求参数", "phone", req.Phone)
	resp, err := h.svc.Register(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		logger.SugarLogger.Debugw("Register: 注册失败", "error", err.Error(), "phone", req.Phone)
		response.Error(c, http.StatusConflict, 10014, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("Login: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	logger.SugarLogger.Debugw("Login: 请求参数", "phone", req.Phone)
	resp, err := h.svc.Login(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		logger.SugarLogger.Debugw("Login: 登录失败", "error", err.Error(), "phone", req.Phone)
		if strings.Contains(err.Error(), "MEMBER_ROLE_FORBIDDEN") {
			response.Error(c, http.StatusForbidden, 10010, "您的账号为员工账号，请使用员工端微信小程序登录")
			return
		}
		response.Error(c, http.StatusBadRequest, 10003, err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("Refresh: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	resp, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		logger.SugarLogger.Debugw("Refresh: 刷新失败", "error", err.Error())
		response.Unauthorized(c, "刷新失败: "+err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")[7:]
	var req RefreshRequest
	_ = c.ShouldBindJSON(&req)
	logger.SugarLogger.Debugw("Logout: 请求", "accessToken", accessToken[:min(20, len(accessToken))]+"...")
	if err := h.svc.Logout(c.Request.Context(), accessToken, req.RefreshToken); err != nil {
		logger.SugarLogger.Debugw("Logout: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, 10004, "退出失败")
		return
	}
	response.Success(c, gin.H{"message": "已退出登录"})
}

func (h *Handler) CompleteOnboarding(c *gin.Context) {
	var req CompleteOnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("CompleteOnboarding: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("CompleteOnboarding: 请求参数", "user_id", userID, "org_name", req.Name)
	resp, err := h.svc.CompleteOnboarding(c.Request.Context(), userID, &req)
	if err != nil {
		logger.SugarLogger.Debugw("CompleteOnboarding: 失败", "error", err.Error(), "user_id", userID)
		response.Error(c, http.StatusInternalServerError, 10005, "企业信息录入失败")
		return
	}
	response.Success(c, resp)
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
	logger.SugarLogger.Debugw("ListSubAccounts: 查询", "org_id", orgID, "page", page, "page_size", pageSize)
	users, total, err := h.svc.ListSubAccounts(c.Request.Context(), orgID, page, pageSize)
	if err != nil {
		logger.SugarLogger.Debugw("ListSubAccounts: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 10006, "查询用户列表失败")
		return
	}
	response.PageSuccess(c, users, total, page, pageSize)
}

func (h *Handler) CreateSubAccount(c *gin.Context) {
	var req CreateSubAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("CreateSubAccount: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("CreateSubAccount: 请求", "org_id", orgID, "phone", req.Phone, "role", req.Role)
	if err := h.svc.CreateSubAccount(c.Request.Context(), orgID, &req); err != nil {
		logger.SugarLogger.Debugw("CreateSubAccount: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusBadRequest, 10007, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "子账号创建成功"})
}

func (h *Handler) UpdateSubAccountRole(c *gin.Context) {
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdateSubAccountRole: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	orgID := c.GetInt64("org_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.SugarLogger.Debugw("UpdateSubAccountRole: 无效ID", "param", c.Param("id"))
		response.BadRequest(c, "无效的用户ID")
		return
	}
	logger.SugarLogger.Debugw("UpdateSubAccountRole: 请求", "org_id", orgID, "target_id", targetID, "new_role", req.Role)
	if err := h.svc.UpdateSubAccountRole(c.Request.Context(), orgID, targetID, req.Role); err != nil {
		logger.SugarLogger.Debugw("UpdateSubAccountRole: 失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, 10008, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "角色更新成功"})
}

func (h *Handler) DeleteSubAccount(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.SugarLogger.Debugw("DeleteSubAccount: 无效ID", "param", c.Param("id"))
		response.BadRequest(c, "无效的用户ID")
		return
	}
	logger.SugarLogger.Debugw("DeleteSubAccount: 请求", "org_id", orgID, "target_id", targetID)
	if err := h.svc.DeleteSubAccount(c.Request.Context(), orgID, targetID); err != nil {
		logger.SugarLogger.Debugw("DeleteSubAccount: 失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, 10009, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "用户已删除"})
}

func (h *Handler) LoginPassword(c *gin.Context) {
	var req PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("LoginPassword: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	logger.SugarLogger.Debugw("LoginPassword: 请求", "phone", req.Phone)
	resp, err := h.svc.LoginPassword(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		logger.SugarLogger.Debugw("LoginPassword: 登录失败", "error", err.Error(), "phone", req.Phone)
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

	logger.SugarLogger.Debugw("GetMe: 请求", "user_id", userID, "org_id", orgID)
	if userID == 0 && orgID == 0 {
		logger.SugarLogger.Warnw("GetMe: user_id 为 0", "user_id", userID)
		response.Error(c, http.StatusInternalServerError, 10013, "获取用户信息失败")
		return
	}
	resp, err := h.svc.GetMe(c.Request.Context(), userID, orgID)
	if err != nil {
		if orgID == 0 {
			logger.SugarLogger.Warnw("GetMe: org_id 为 0", "user_id", userID, "org_id", orgID)
			response.Error(c, http.StatusInternalServerError, 10018, "请先完善企业信息")
			return
		}
		logger.SugarLogger.Debugw("GetMe: 失败", "error", err.Error(), "user_id", userID)
		response.Error(c, http.StatusInternalServerError, 10013, "获取用户信息失败")
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("ChangePassword: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("ChangePassword: 请求", "user_id", userID)
	if err := h.svc.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		logger.SugarLogger.Debugw("ChangePassword: 失败", "error", err.Error(), "user_id", userID)
		response.Error(c, http.StatusBadRequest, 10014, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "密码修改成功"})
}

func (h *Handler) UpdateOrg(c *gin.Context) {
	var req UpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdateOrg: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	orgID := c.GetInt64("org_id")
	logger.SugarLogger.Debugw("UpdateOrg: 请求", "org_id", orgID, "name", req.Name)
	if err := h.svc.UpdateOrg(c.Request.Context(), orgID, &req); err != nil {
		logger.SugarLogger.Debugw("UpdateOrg: 失败", "error", err.Error(), "org_id", orgID)
		response.Error(c, http.StatusInternalServerError, 10015, "更新企业信息失败")
		return
	}
	response.Success(c, gin.H{"message": "企业信息更新成功"})
}

func (h *Handler) UpdateAvatar(c *gin.Context) {
	var req UpdateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdateAvatar: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("UpdateAvatar: 请求", "user_id", userID)
	if err := h.svc.UpdateAvatar(c.Request.Context(), userID, req.Avatar); err != nil {
		logger.SugarLogger.Debugw("UpdateAvatar: 失败", "error", err.Error(), "user_id", userID)
		response.Error(c, http.StatusInternalServerError, 10016, "更新头像失败")
		return
	}
	response.Success(c, gin.H{"message": "头像更新成功"})
}

func (h *Handler) UpdateName(c *gin.Context) {
	var req UpdateNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("UpdateName: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("UpdateName: 请求", "user_id", userID, "name", req.Name)
	if err := h.svc.UpdateName(c.Request.Context(), userID, req.Name); err != nil {
		logger.SugarLogger.Debugw("UpdateName: 失败", "error", err.Error(), "user_id", userID)
		response.Error(c, http.StatusInternalServerError, 10017, "更新姓名失败")
		return
	}
	response.Success(c, gin.H{"message": "姓名更新成功"})
}

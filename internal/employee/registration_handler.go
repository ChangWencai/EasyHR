package employee

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
	"github.com/wencai/easyhr/pkg/sms"
)

// RegistrationHandler 员工信息登记 HTTP 端点
type RegistrationHandler struct {
	svc      *RegistrationService
	smsClient *sms.Client
}

// NewRegistrationHandler 创建登记 Handler
func NewRegistrationHandler(svc *RegistrationService, smsClient *sms.Client) *RegistrationHandler {
	return &RegistrationHandler{svc: svc, smsClient: smsClient}
}

// RegisterRoutes 注册路由（公开接口 + 认证接口混合）
func (h *RegistrationHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("")
	authGroup.Use(authMiddleware)

	// 公开接口（无需认证）— 员工填写信息
	rg.GET("/registrations/:token", h.GetRegistrationDetail)
	rg.POST("/registrations/:token/submit", h.SubmitRegistration)
	rg.POST("/registrations/send-sms", h.SendRegistrationSms)

	// 管理接口（需要认证）
	authGroup.POST("/registrations", middleware.RequireRole("owner", "admin"), h.CreateRegistration)
	authGroup.GET("/registrations", h.ListRegistrations)
	authGroup.DELETE("/registrations/:id", middleware.RequireRole("owner", "admin"), h.DeleteRegistration)
}

// CreateRegistration 创建登记表
func (h *RegistrationHandler) CreateRegistration(c *gin.Context) {
	var req CreateRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	resp, err := h.svc.CreateRegistration(orgID, userID, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 20300, err.Error())
		return
	}

	response.Success(c, resp)
}

// ListRegistrations 登记列表
func (h *RegistrationHandler) ListRegistrations(c *gin.Context) {
	var params RegistrationListQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	items, total, err := h.svc.ListRegistrations(orgID, params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 20301, "查询登记列表失败")
		return
	}

	response.PageSuccess(c, items, total, params.Page, params.PageSize)
}

// GetRegistrationDetail 获取登记详情（公开接口）
func (h *RegistrationHandler) GetRegistrationDetail(c *gin.Context) {
	token := c.Param("token")

	detail, err := h.svc.GetRegistrationDetail(token)
	if err != nil {
		if err == ErrRegistrationNotFound {
			response.Error(c, http.StatusNotFound, 20206, "链接无效，请确认链接是否正确")
			return
		}
		if err == ErrRegistrationExpired {
			response.Error(c, http.StatusGone, 20207, "该链接已过期，请联系管理员重新发送")
			return
		}
		if err == ErrRegistrationAlreadyUsed {
			response.Error(c, http.StatusBadRequest, 20302, "该登记已提交")
			return
		}
		response.Error(c, http.StatusInternalServerError, 20303, "查询登记详情失败")
		return
	}

	response.Success(c, detail)
}

// SubmitRegistration 员工提交登记信息（公开接口）
func (h *RegistrationHandler) SubmitRegistration(c *gin.Context) {
	token := c.Param("token")

	var req SubmitRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.SubmitRegistration(token, &req); err != nil {
		if err == ErrRegistrationNotFound {
			response.Error(c, http.StatusNotFound, 20206, "链接无效，请确认链接是否正确")
			return
		}
		if err == ErrRegistrationExpired {
			response.Error(c, http.StatusGone, 20207, "该链接已过期，请联系管理员重新发送")
			return
		}
		if err == ErrRegistrationAlreadyUsed {
			response.Error(c, http.StatusBadRequest, 20304, "该登记已提交")
			return
		}
		response.Error(c, http.StatusBadRequest, 20305, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "信息提交成功，感谢您的配合"})
}

// DeleteRegistration 删除登记表
func (h *RegistrationHandler) DeleteRegistration(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的登记ID")
		return
	}

	if err := h.svc.DeleteRegistration(orgID, id); err != nil {
		if err == ErrRegistrationNotFound {
			response.Error(c, http.StatusNotFound, 20306, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 20307, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "登记表已删除"})
}

// SendRegistrationSms 发送登记链接短信（公开接口）
func (h *RegistrationHandler) SendRegistrationSms(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	registrationUrl := fmt.Sprintf("%s/#/register/%s", c.Request.Host, req.Token)
	if err := h.smsClient.SendTemplateSMS(c.Request.Context(), req.Phone, fmt.Sprintf(`{"url":"%s"}`, registrationUrl)); err != nil {
		response.Error(c, http.StatusInternalServerError, 20308, "短信发送失败")
		return
	}

	response.Success(c, gin.H{"message": "短信已发送"})
}

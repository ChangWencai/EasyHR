package salary

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/response"
)

// UnlockHandler 解锁 HTTP 端点
type UnlockHandler struct {
	unlockSvc *UnlockService
}

// NewUnlockHandler 创建解锁 Handler
func NewUnlockHandler(unlockSvc *UnlockService) *UnlockHandler {
	return &UnlockHandler{unlockSvc: unlockSvc}
}

// RegisterRoutes 注册解锁路由
func (h *UnlockHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	unlock := rg.Group("/salary", authMiddleware)
	{
		unlock.POST("/unlock", middleware.RequireRole("owner", "admin"), h.Unlock)
		unlock.POST("/unlock/send-code", middleware.RequireRole("owner", "admin"), h.SendCode)
	}
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// SendCode 发送解锁验证码
func (h *UnlockHandler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	logger.SugarLogger.Debugw("SendCode: 发送解锁验证码", "phone", req.Phone)

	if err := h.unlockSvc.SendUnlockCode(c.Request.Context(), req.Phone); err != nil {
		logger.SugarLogger.Debugw("SendCode: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "验证码已发送"})
}

// UnlockRequest 解锁请求
type UnlockPayrollRequest struct {
	RecordID int64  `json:"record_id" binding:"required"`
	SMSCode  string `json:"sms_code" binding:"required,len=6"`
	Phone    string `json:"phone" binding:"required"`
}

// Unlock 解锁工资记录
func (h *UnlockHandler) Unlock(c *gin.Context) {
	var req UnlockPayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")
	logger.SugarLogger.Debugw("Unlock: 解锁工资记录", "org_id", orgID, "user_id", userID, "record_id", req.RecordID)

	if err := h.unlockSvc.UnlockPayroll(orgID, userID, req.RecordID, req.SMSCode, req.Phone); err != nil {
		if err == ErrUnlockInvalidCode {
			response.Error(c, http.StatusBadRequest, CodeTemplateConfig, "验证码错误")
			return
		}
		if err == ErrUnlockRecordNotLocked {
			response.Error(c, http.StatusBadRequest, CodeInvalidStatus, "该记录未锁定，无需解锁")
			return
		}
		if err == ErrUnlockCodeExpired {
			response.Error(c, http.StatusBadRequest, CodeTemplateConfig, "验证码已过期，请重新获取")
			return
		}
		logger.SugarLogger.Debugw("Unlock: 失败", "error", err.Error())
		response.Error(c, http.StatusBadRequest, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "解锁成功，请重新编辑并确认"})
}

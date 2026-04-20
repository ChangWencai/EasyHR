package salary

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wencai/easyhr/internal/common/logger"
	"github.com/wencai/easyhr/internal/common/response"
)

// SlipSendHandler 工资条发送 HTTP 端点
type SlipSendHandler struct {
	slipSendSvc *SlipSendService
}

// NewSlipSendHandler 创建工资条发送 Handler
func NewSlipSendHandler(slipSendSvc *SlipSendService) *SlipSendHandler {
	return &SlipSendHandler{slipSendSvc: slipSendSvc}
}

// RegisterRoutes 注册路由
func (h *SlipSendHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	slip := rg.Group("/salary/slip", authMiddleware)
	{
		slip.POST("/send-all", h.SendAllSlips)
		slip.GET("/logs", h.GetSlipLogs) // 查询发送日志
	}
}

// SendAllSlips 批量发送工资条（全员或选定）
func (h *SlipSendHandler) SendAllSlips(c *gin.Context) {
	var req SendAllSlipsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.SugarLogger.Debugw("SendAllSlips: 参数错误", "error", err.Error())
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	orgID := c.GetInt64("org_id")
	userID := c.GetInt64("user_id")

	logger.SugarLogger.Debugw("SendAllSlips: 调用",
		"org_id", orgID, "user_id", userID,
		"year", req.Year, "month", req.Month,
		"employee_ids", req.EmployeeIDs, "channel", req.Channel)

	if err := h.slipSendSvc.SendAllSlips(orgID, userID, req.Year, req.Month, req.EmployeeIDs, req.Channel); err != nil {
		logger.SugarLogger.Debugw("SendAllSlips: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "工资条发送任务已入队，请稍后在发送记录中查看进度",
	})
}

// GetSlipLogs 查询工资条发送日志（含员工确认状态 D-13-11）
func (h *SlipSendHandler) GetSlipLogs(c *gin.Context) {
	orgID := c.GetInt64("org_id")

	yearStr := c.Query("year")
	monthStr := c.Query("month")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	var year, month int
	var err error
	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil || year < 2000 || year > 2100 {
			response.BadRequest(c, "year 参数错误")
			return
		}
	}
	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			response.BadRequest(c, "month 参数错误")
			return
		}
	}

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logger.SugarLogger.Debugw("GetSlipLogs: 查询", "org_id", orgID, "year", year, "month", month)

	// 查询工资条发送日志（含 confirmed_at 联查）
	logs, err := h.slipSendSvc.GetSlipLogsWithConfirmation(orgID, year, month, page, pageSize)
	if err != nil {
		logger.SugarLogger.Debugw("GetSlipLogs: 失败", "error", err.Error())
		response.Error(c, http.StatusInternalServerError, CodeTemplateConfig, "查询失败")
		return
	}

	response.Success(c, gin.H{
		"logs":      logs,
		"page":      page,
		"page_size": pageSize,
	})
}

// SendAllSlipsRequest 批量发送请求
type SendAllSlipsRequest struct {
	Year        int     `json:"year" binding:"required,min=2000,max=2100"`
	Month       int     `json:"month" binding:"required,min=1,max=12"`
	EmployeeIDs []int64 `json:"employee_ids"` // 空=全员
	Channel     string  `json:"channel"`       // miniapp/sms/h5，默认 miniapp
}
